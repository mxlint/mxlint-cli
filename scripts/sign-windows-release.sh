#!/usr/bin/env bash
# Sign Windows release .exe assets locally with Certum SimplySign + osslsigncode,
# upload them back to the GitHub release under a *-signed.exe name, then
# remove the original unsigned .exe assets from the release.
#
# Prerequisites:
#   - SimplySign Desktop installed and LOGGED IN (cloud card mounted; ~2h session)
#   - brew install osslsigncode libp11 opensc gh
#   - gh authenticated to the mxlint-cli repo
#
# Required env:
#   CERTUM_SIGN_CERT   PEM with the full chain: leaf FIRST, then Certum
#                      intermediate(s). Leaf-only often fails verification.
#   CERTUM_KEY_ID      PKCS#11 key id (hex, with or without colons) OR a full
#                      pkcs11: URI. On SimplySign the cert shares the key's
#                      CKA_ID; discover it WITHOUT --login (cloud card has no PIN):
#                        pkcs11-tool --module "$CERTUM_PKCS11_MODULE" \
#                          --list-objects --type cert
#
# Optional env:
#   CERTUM_PKCS11_MODULE  SimplySign PKCS#11 .dylib (auto-detected on macOS)
#   CERTUM_TS_URL         timestamp URL (default http://time.certum.pl/)
#   MXLINT_REPO           owner/name override (default: inferred via gh)
#   OPENSSL_ENGINES       dir containing libp11's pkcs11.dylib (auto-detected)
#
# Usage:
#   ./scripts/sign-windows-release.sh v3.17.0
#
# Downloads:
#   mxlint-<tag>-windows-amd64.exe
#   mxlint-<tag>-windows-arm64.exe
# Uploads:
#   mxlint-<tag>-windows-amd64-signed.exe
#   mxlint-<tag>-windows-arm64-signed.exe

set -euo pipefail

TAG="${1:-}"
if [[ -z "$TAG" ]]; then
  echo "Usage: $0 <version-tag>   e.g. $0 v3.17.0" >&2
  exit 1
fi

TS_URL="${CERTUM_TS_URL:-http://time.certum.pl/}"
ARCHES=(amd64 arm64)

die() { echo "Error: $*" >&2; exit 1; }

for tool in gh osslsigncode; do
  command -v "$tool" >/dev/null 2>&1 || die "'$tool' is required (brew install $tool)"
done

: "${CERTUM_SIGN_CERT:?Set CERTUM_SIGN_CERT to the code-signing certificate PEM (full chain)}"
: "${CERTUM_KEY_ID:?Set CERTUM_KEY_ID to the PKCS#11 key id / pkcs11: URI}"
[[ -f "$CERTUM_SIGN_CERT" ]] || die "cert PEM not found: $CERTUM_SIGN_CERT"

# --- PKCS#11 module (SimplySign) ---------------------------------------------
if [[ -z "${CERTUM_PKCS11_MODULE:-}" ]]; then
  for cand in \
    /usr/local/lib/libSimplySignPKCS.dylib \
    /Applications/proCertumSmartSign.app/Contents/MacOS/libSimplySignPKCS.dylib; do
    if [[ -f "$cand" ]]; then
      CERTUM_PKCS11_MODULE="$cand"
      break
    fi
  done
fi
[[ -n "${CERTUM_PKCS11_MODULE:-}" ]] || die "SimplySign PKCS#11 module not found; set CERTUM_PKCS11_MODULE"
[[ -f "$CERTUM_PKCS11_MODULE" ]] || die "PKCS#11 module not found: $CERTUM_PKCS11_MODULE"

# Resolve symlink — some PKCS#11 tooling rejects the /usr/local/lib symlink.
if [[ -L "$CERTUM_PKCS11_MODULE" ]]; then
  _link="$(readlink "$CERTUM_PKCS11_MODULE")"
  case "$_link" in
    /*) CERTUM_PKCS11_MODULE="$_link" ;;
    *) CERTUM_PKCS11_MODULE="$(cd "$(dirname "$CERTUM_PKCS11_MODULE")" && pwd)/$_link" ;;
  esac
  echo "Resolved PKCS#11 module -> $CERTUM_PKCS11_MODULE"
fi

# --- OpenSSL pkcs11 engine (libp11) ------------------------------------------
# osslsigncode talks to PKCS#11 through OpenSSL's pkcs11 ENGINE. On OpenSSL 3
# that engine comes from libp11, not OpenSSL itself.
if [[ -z "${OPENSSL_ENGINES:-}" ]]; then
  for d in \
    /opt/homebrew/lib/engines-3 \
    /usr/local/lib/engines-3 \
    "$(brew --prefix libp11 2>/dev/null)/lib/engines-3" \
    /usr/lib/engines-3; do
    [[ -n "$d" ]] || continue
    if [[ -f "$d/pkcs11.dylib" || -f "$d/pkcs11.so" ]]; then
      export OPENSSL_ENGINES="$d"
      break
    fi
  done
fi
if [[ -z "${OPENSSL_ENGINES:-}" ]]; then
  die "PKCS#11 engine (libp11) not found.
  Install: brew install libp11
  Or set OPENSSL_ENGINES to the directory containing pkcs11.dylib"
fi
echo "Using PKCS#11 engine dir: $OPENSSL_ENGINES"
echo "Using PKCS#11 module:     $CERTUM_PKCS11_MODULE"

# --- Key URI -----------------------------------------------------------------
case "$CERTUM_KEY_ID" in
  pkcs11:*) KEY_URI="$CERTUM_KEY_ID" ;;
  *)
    _hex="$(printf '%s' "$CERTUM_KEY_ID" | tr -d ':[:space:]')"
    _enc="$(printf '%s' "$_hex" | sed 's/\(..\)/%\1/g')"
    KEY_URI="pkcs11:id=${_enc};type=private"
    ;;
esac
echo "Using key URI: $KEY_URI"

# Capture repo before leaving the checkout (temp workdir has no git remote).
REPO="${MXLINT_REPO:-$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null || true)}"
[[ -n "$REPO" ]] || die "could not infer GitHub repo; set MXLINT_REPO=owner/name"
echo "Target repo: $REPO  tag: $TAG"

WORK="$(mktemp -d "${TMPDIR:-/tmp}/mxlint-sign.XXXXXX")"
cleanup() { rm -rf "$WORK"; }
trap cleanup EXIT

SIGNED_UPLOADS=()
UNSIGNED_ASSETS=()

for arch in "${ARCHES[@]}"; do
  unsigned="mxlint-${TAG}-windows-${arch}.exe"
  signed="mxlint-${TAG}-windows-${arch}-signed.exe"

  echo ""
  echo "=== $arch ==="
  echo "Downloading $unsigned ..."
  gh release download "$TAG" --repo "$REPO" --pattern "$unsigned" --dir "$WORK" --clobber
  [[ -f "$WORK/$unsigned" ]] || die "asset not found on release $TAG: $unsigned"

  echo "Signing -> $signed (SimplySign must be logged in) ..."
  # -t (legacy Authenticode timestamp) is the Certum-proven form for
  # time.certum.pl. RFC3161 (-ts) has been reported to fail there.
  osslsigncode sign \
    -pkcs11module "$CERTUM_PKCS11_MODULE" \
    -certs "$CERTUM_SIGN_CERT" \
    -key "$KEY_URI" \
    -h sha256 \
    -t "$TS_URL" \
    -in "$WORK/$unsigned" \
    -out "$WORK/$signed"

  echo "Verifying $signed ..."
  osslsigncode verify "$WORK/$signed"

  SIGNED_UPLOADS+=("$WORK/$signed")
  UNSIGNED_ASSETS+=("$unsigned")
done

echo ""
echo "Uploading signed assets to $TAG ..."
gh release upload "$TAG" "${SIGNED_UPLOADS[@]}" --repo "$REPO" --clobber

echo ""
echo "Removing unsigned assets from $TAG ..."
for asset in "${UNSIGNED_ASSETS[@]}"; do
  gh release delete-asset "$TAG" "$asset" --repo "$REPO" --yes
done

echo ""
echo "Done. Uploaded:"
for f in "${SIGNED_UPLOADS[@]}"; do
  echo "  - $(basename "$f")"
done
