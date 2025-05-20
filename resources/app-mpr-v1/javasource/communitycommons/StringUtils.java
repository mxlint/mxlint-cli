package communitycommons;

import com.mendix.core.Core;
import com.mendix.systemwideinterfaces.core.IContext;
import com.mendix.systemwideinterfaces.core.IMendixObject;
import communitycommons.proxies.SanitizerPolicy;

import static communitycommons.proxies.SanitizerPolicy.BLOCKS;
import static communitycommons.proxies.SanitizerPolicy.FORMATTING;
import static communitycommons.proxies.SanitizerPolicy.IMAGES;
import static communitycommons.proxies.SanitizerPolicy.LINKS;
import static communitycommons.proxies.SanitizerPolicy.STYLES;
import static communitycommons.proxies.SanitizerPolicy.TABLES;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.StringReader;
import java.io.UnsupportedEncodingException;
import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;
import java.security.*;
import java.text.Normalizer;
import java.util.*;
import java.util.AbstractMap.SimpleEntry;
import java.util.function.Function;
import java.util.regex.MatchResult;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;
import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import javax.swing.text.MutableAttributeSet;
import javax.swing.text.html.HTML;
import javax.swing.text.html.HTMLEditorKit;
import javax.swing.text.html.parser.ParserDelegator;

import org.apache.commons.io.IOUtils;
import org.apache.commons.text.StringEscapeUtils;
import org.owasp.html.PolicyFactory;
import org.owasp.html.Sanitizers;
import system.proxies.FileDocument;

public class StringUtils {

	private static final Random RANDOM = new SecureRandom();

	private static final String UPPERCASE_ALPHA = stringRange('A', 'Z');
	private static final String LOWERCASE_ALPHA = stringRange('a', 'z');
	private static final String DIGITS = stringRange('0', '9');
	// Used in tests as well
	static final String SPECIAL = stringRange('!', '/');
	private static final String ALPHANUMERIC = UPPERCASE_ALPHA + LOWERCASE_ALPHA + DIGITS;

	static final Map<String, PolicyFactory> SANITIZER_POLICIES =
		Map.ofEntries(
			new SimpleEntry<>(BLOCKS.name(), Sanitizers.BLOCKS),
			new SimpleEntry<>(FORMATTING.name(), Sanitizers.FORMATTING),
			new SimpleEntry<>(IMAGES.name(), Sanitizers.IMAGES),
			new SimpleEntry<>(LINKS.name(), Sanitizers.LINKS),
			new SimpleEntry<>(STYLES.name(), Sanitizers.STYLES),
			new SimpleEntry<>(TABLES.name(), Sanitizers.TABLES)
		);

	public static final String HASH_ALGORITHM = "SHA-256";

	public static String hash(String value, int length) throws NoSuchAlgorithmException, DigestException {
		byte[] inBytes = value.getBytes(StandardCharsets.UTF_8);
		byte[] outBytes = new byte[length];

		MessageDigest alg = MessageDigest.getInstance(HASH_ALGORITHM);
		alg.update(inBytes);

		alg.digest(outBytes, 0, length);

		StringBuilder hexString = new StringBuilder();
		for (int i = 0; i < outBytes.length; i++) {
			String hex = Integer.toHexString(0xff & outBytes[i]);
			if (hex.length() == 1) {
				hexString.append('0');
			}
			hexString.append(hex);
		}

		return hexString.toString();
	}

	/**
	 * The default replaceAll microflow function doesn't support capture variables such as $1, $2
	 * etc. so for that reason we do not deprecate this method.
	 *
	 * @param haystack    The string to replace patterns in
	 * @param needleRegex The regular expression pattern
	 * @param replacement The string that should come in place of the pattern matches.
	 * @return The resulting string, where all matches have been replaced by the replacement.
	 */
	public static String regexReplaceAll(String haystack, String needleRegex,
										 String replacement) {
		Pattern pattern = Pattern.compile(needleRegex);
		Matcher matcher = pattern.matcher(haystack);
		return matcher.replaceAll(replacement);
	}

	public static String leftPad(String value, Long amount, String fillCharacter) {
		if (fillCharacter == null || fillCharacter.length() == 0) {
			return org.apache.commons.lang3.StringUtils.leftPad(value, amount.intValue(), " ");
		}
		return org.apache.commons.lang3.StringUtils.leftPad(value, amount.intValue(), fillCharacter);
	}

	public static String rightPad(String value, Long amount, String fillCharacter) {
		if (fillCharacter == null || fillCharacter.length() == 0) {
			return org.apache.commons.lang3.StringUtils.rightPad(value, amount.intValue(), " ");
		}
		return org.apache.commons.lang3.StringUtils.rightPad(value, amount.intValue(), fillCharacter);
	}

	public static String randomString(int length) {
		return randomStringFromCharArray(length, ALPHANUMERIC.toCharArray());
	}

	public static String substituteTemplate(final IContext context, String template,
											final IMendixObject substitute, final boolean HTMLEncode, final String datetimeformat) {
		return regexReplaceAll(template, "\\{(@)?([\\w./]+)\\}", (MatchResult match) -> {
			String value;
			String path = match.group(2);
			if (match.group(1) != null) {
				value = String.valueOf(Core.getConfiguration().getConstantValue(path));
			} else {
				try {
					value = ORM.getValueOfPath(context, substitute, path, datetimeformat);
				} catch (Exception e) {
					throw new RuntimeException(e);
				}
			}
			return HTMLEncode ? HTMLEncode(value) : value;
		});
	}

	public static String regexReplaceAll(String source, String regexString, Function<MatchResult, String> replaceFunction) {
		if (source == null || source.trim().isEmpty()) // avoid NPE's, save CPU
		{
			return "";
		}

		StringBuffer resultString = new StringBuffer();
		Pattern regex = Pattern.compile(regexString);
		Matcher regexMatcher = regex.matcher(source);

		while (regexMatcher.find()) {
			MatchResult match = regexMatcher.toMatchResult();
			String value = replaceFunction.apply(match);
			regexMatcher.appendReplacement(resultString, Matcher.quoteReplacement(value));
		}
		regexMatcher.appendTail(resultString);

		return resultString.toString();
	}

	public static String HTMLEncode(String value) {
		return StringEscapeUtils.escapeHtml4(value);
	}

	public static String randomHash() {
		return UUID.randomUUID().toString();
	}

	public static String base64Decode(String encoded) {
		if (encoded == null) {
			return null;
		}
		return new String(Base64.getDecoder().decode(encoded.getBytes()));
	}

	public static void base64DecodeToFile(IContext context, String encoded, FileDocument targetFile) throws Exception {
		if (targetFile == null) {
			throw new IllegalArgumentException("Source file is null");
		}
		if (encoded == null) {
			throw new IllegalArgumentException("Source data is null");
		}

		byte[] decoded = Base64.getDecoder().decode(encoded.getBytes());

		try (ByteArrayInputStream bais = new ByteArrayInputStream(decoded)) {
			Core.storeFileDocumentContent(context, targetFile.getMendixObject(), bais);
		}
	}

	public static String base64Encode(String value) {
		if (value == null) {
			return null;
		}
		return Base64.getEncoder().encodeToString(value.getBytes());
	}

	public static String base64EncodeFile(IContext context, FileDocument file) throws IOException {
		if (file == null) {
			throw new IllegalArgumentException("Source file is null");
		}
		if (!file.getHasContents()) {
			throw new IllegalArgumentException("Source file has no contents!");
		}

		try (InputStream f = Core.getFileDocumentContent(context, file.getMendixObject())) {
			return Base64.getEncoder().encodeToString(IOUtils.toByteArray(f));
		}
	}

	public static String stringFromFile(IContext context, FileDocument source) throws IOException {
		return stringFromFile(context, source, StandardCharsets.UTF_8);
	}

	public static String stringFromFile(IContext context, FileDocument source, Charset charset) throws IOException {
		if (source == null) {
			return null;
		}
		try (InputStream f = Core.getFileDocumentContent(context, source.getMendixObject())) {
			return IOUtils.toString(f, charset);
		}
	}

	public static void stringToFile(IContext context, String value, FileDocument destination) throws IOException {
		stringToFile(context, value, destination, StandardCharsets.UTF_8);
	}

	public static void stringToFile(IContext context, String value, FileDocument destination, Charset charset) throws IOException {
		if (destination == null) {
			throw new IllegalArgumentException("Destination file is null");
		}
		if (value == null) {
			throw new IllegalArgumentException("Value to write is null");
		}

		try (InputStream is = IOUtils.toInputStream(value, charset)) {
			Core.storeFileDocumentContent(context, destination.getMendixObject(), is);
		}
	}

	public static String HTMLToPlainText(String html) throws IOException {
		if (html == null) {
			return "";
		}
		final StringBuilder result = new StringBuilder();

		HTMLEditorKit.ParserCallback callback = new HTMLEditorKit.ParserCallback() {
			@Override
			public void handleText(char[] data, int pos) {
				result.append(data); //TODO: needds to be html entity decode?
			}

			@Override
			public void handleComment(char[] data, int pos) {
				//Do nothing
			}

			@Override
			public void handleError(String errorMsg, int pos) {
				//Do nothing
			}

			@Override
			public void handleSimpleTag(HTML.Tag tag, MutableAttributeSet a, int pos) {
				if (tag == HTML.Tag.BR) {
					result.append("\r\n");
				}
			}

			@Override
			public void handleEndTag(HTML.Tag tag, int pos) {
				if (tag == HTML.Tag.P) {
					result.append("\r\n");
				}
			}
		};

		new ParserDelegator().parse(new StringReader(html), callback, true);

		return result.toString();
	}

	/**
	 * Returns a random strong password containing a specified minimum number of uppercase, digits
	 * and the exact number of special characters.
	 *
	 * @param minLen        Minimum length
	 * @param maxLen        Maximum length
	 * @param noOfCAPSAlpha Minimum number of capitals
	 * @param noOfDigits    Minimum number of digits
	 * @param noOfSplChars  Exact number of special characters
	 * @deprecated          Use the overload randomStrongPassword instead
	 */
	@Deprecated
	public static String randomStrongPassword(int minLen, int maxLen, int noOfCAPSAlpha, int noOfDigits, int noOfSplChars) {
		if (minLen > maxLen) {
			throw new IllegalArgumentException("Min. Length > Max. Length!");
		}
		if ((noOfCAPSAlpha + noOfDigits + noOfSplChars) > minLen) {
			throw new IllegalArgumentException("Min. Length should be atleast sum of (CAPS, DIGITS, SPL CHARS) Length!");
		}
		return generateCommonLangPassword(minLen, maxLen, noOfCAPSAlpha, 0, noOfDigits, noOfSplChars);
	}

	/**
	 * Returns a random strong password containing a specified minimum number of uppercase, lowercase, digits
	 * and the exact number of special characters.
	 *
	 * @param minLen             Minimum length
	 * @param maxLen             Maximum length
	 * @param noOfCAPSAlpha      Minimum number of capitals
	 * @param noOfLowercaseAlpha Minimum number of lowercase letters
	 * @param noOfDigits         Minimum number of digits
	 * @param noOfSplChars       Exact number of special characters
	 */
	public static String randomStrongPassword(int minLen, int maxLen, int noOfCAPSAlpha, int noOfLowercaseAlpha, int noOfDigits, int noOfSplChars) {
		if (minLen > maxLen) {
			throw new IllegalArgumentException("Min. Length > Max. Length!");
		}
		if ((noOfCAPSAlpha + noOfLowercaseAlpha + noOfDigits + noOfSplChars) > minLen) {
			throw new IllegalArgumentException("Min. Length should be atleast sum of (CAPS, LOWER, DIGITS, SPL CHARS) Length!");
		}
		return generateCommonLangPassword(minLen, maxLen, noOfCAPSAlpha, noOfLowercaseAlpha, noOfDigits, noOfSplChars);
	}

	// See https://www.baeldung.com/java-generate-secure-password
	// Implementation inspired by https://github.com/eugenp/tutorials/tree/master/core-java-modules/core-java-string-apis (under MIT license)
	private static String generateCommonLangPassword(int minLen, int maxLen, int noOfCapsAlpha, int noOfLowercaseAlpha, int noOfDigits, int noOfSplChars) {
		String upperCaseLetters = randomStringFromCharArray(noOfCapsAlpha, UPPERCASE_ALPHA.toCharArray());
		String lowerCaseLetters = randomStringFromCharArray(noOfLowercaseAlpha, LOWERCASE_ALPHA.toCharArray());
		String numbers = randomStringFromCharArray(noOfDigits, DIGITS.toCharArray());
		String specialChar = randomStringFromCharArray(noOfSplChars, SPECIAL.toCharArray());

		final int fixedNumber = noOfCapsAlpha + noOfLowercaseAlpha + noOfDigits + noOfSplChars;
		final int lowerBound = minLen - fixedNumber;
		final int upperBound = maxLen - fixedNumber;
		String totalChars = randomStringFromCharArray(lowerBound, upperBound, ALPHANUMERIC.toCharArray());

		String combinedChars = upperCaseLetters
			.concat(lowerCaseLetters)
			.concat(numbers)
			.concat(specialChar)
			.concat(totalChars);
		List<Character> pwdChars = combinedChars.chars()
			.mapToObj(c -> (char) c)
			.collect(Collectors.toList());
		Collections.shuffle(pwdChars);
		String password = pwdChars.stream()
			.collect(StringBuilder::new, StringBuilder::append, StringBuilder::append)
			.toString();
		return password;
	}

	/**
	 * Generate a secure random string using the given array of characters, of which the resulting
	 * string will be composed of.
	 *
	 * @param count        The length of the random string.
	 * @param allowedChars The characters used for constructing the random string.
	 * @return A random string.
	 * @throws IllegalArgumentException if <code>count</code> is negative or <code>allowedChars</code> is null or empty.
	 */
	private static String randomStringFromCharArray(int count, final char[] allowedChars) {
		if (count == 0)
			return "";
		if (count < 0)
			throw new IllegalArgumentException("The requested length for the random string was negative: " + count);
		if (allowedChars == null)
			throw new IllegalArgumentException("The char array 'allowedChars' cannot be null.");
		if (allowedChars.length == 0)
			throw new IllegalArgumentException("The char array 'allowedChars' cannot be empty.");

		StringBuilder builder = new StringBuilder();

		while (count-- > 0) {
			int index = RANDOM.nextInt(allowedChars.length);
			builder.append(allowedChars[index]);
		}

		return builder.toString();
	}

	/**
	 * Generate a random string with a random length between <code>minLengthBound</code> and <code>maxLengthBound</code> (inclusive),
	 * using the given set of allowed characters.
	 *
	 * @param minLengthBound The lower bound for the random length of the string.
	 * @param maxLengthBound The upper bound for the random length of the string.
	 * @param allowedChars   An array of characters of which the resulting string will be made up of.
	 * @return A random string with a length between <code>minLengthBound</code> and <code>maxLengthBound</code>.
	 * @throws IllegalArgumentException if <code>minLengthBound</code> is larger than <code>maxLengthBound</code>.
	 */
	private static String randomStringFromCharArray(int minLengthBound, int maxLengthBound, final char[] allowedChars) {
		if (minLengthBound == maxLengthBound)
			return randomStringFromCharArray(minLengthBound, allowedChars);
		if (minLengthBound > maxLengthBound)
			throw new IllegalArgumentException("The minimum bound (" + minLengthBound + ") was larger than the maximum bound (" + maxLengthBound + ".");
		final int randomLength = minLengthBound + RANDOM.nextInt(maxLengthBound - minLengthBound + 1); // add one to make the range inclusive.
		return randomStringFromCharArray(randomLength, allowedChars);
	}

	/**
	 * Produces a 'range' string starting from the <code>begin</code> character up to
	 * the <code>end</code> character (inclusive range). For example, for the range (a-z),
	 * this method will generate the lowercase alphabet.
	 *
	 * @param begin The starting point of the string.
	 * @param end   The ending point of the string.
	 * @return A string from <code>begin</code> to <code>end</code> (inclusive range).
	 * @throws IllegalArgumentException if the <code>begin</code> character has a higher code point than the <code>end</code> character.
	 */
	private static String stringRange(char begin, char end) {
		if (begin > end) {
			throw new IllegalArgumentException("The 'begin' character cannot be larger than the 'end' character.");
		}

		StringBuilder builder = new StringBuilder();
		while (begin <= end)
			builder.append(begin++);
		return builder.toString();
	}

	private static byte[] generateHmacSha256Bytes(String key, String valueToEncrypt) throws UnsupportedEncodingException, IllegalStateException, InvalidKeyException, NoSuchAlgorithmException {
		SecretKeySpec secretKey = new SecretKeySpec(key.getBytes("UTF-8"), "HmacSHA256");
		Mac mac = Mac.getInstance("HmacSHA256");
		mac.init(secretKey);
		mac.update(valueToEncrypt.getBytes("UTF-8"));
		byte[] hmacData = mac.doFinal();

		return hmacData;
	}

	public static String generateHmacSha256(String key, String valueToEncrypt) {
		try {
			byte[] hash = generateHmacSha256Bytes(key, valueToEncrypt);
			StringBuilder result = new StringBuilder();
			for (byte b : hash) {
				result.append(String.format("%02x", b));
			}
			return result.toString();
		} catch (UnsupportedEncodingException | IllegalStateException | InvalidKeyException | NoSuchAlgorithmException e) {
			throw new RuntimeException("CommunityCommons::generateHmacSha256::Unable to encode: " + e.getMessage(), e);
		}
	}

	public static String generateHmacSha256Hash(String key, String valueToEncrypt) {
		try {
			return Base64.getEncoder().encodeToString(generateHmacSha256Bytes(key, valueToEncrypt));
		} catch (UnsupportedEncodingException | IllegalStateException | InvalidKeyException | NoSuchAlgorithmException e) {
			throw new RuntimeException("CommunityCommons::generateHmacSha256Hash::Unable to encode: " + e.getMessage(), e);
		}
	}

	public static String escapeHTML(String input) {
		return input.replace("&", "&amp;")
			.replace("<", "&lt;")
			.replace(">", "&gt;")
			.replace("\"", "&quot;")
			.replace("'", "&#39;");// notice this one: for xml "&#39;" would be "&apos;" (http://blogs.msdn.com/b/kirillosenkov/archive/2010/03/19/apos-is-in-xml-in-html-use-39.aspx)
		// OWASP also advises to escape "/" but give no convincing reason why: https://www.owasp.org/index.php/XSS_%28Cross_Site_Scripting%29_Prevention_Cheat_Sheet
	}

	public static String regexQuote(String unquotedLiteral) {
		return Pattern.quote(unquotedLiteral);
	}

	public static String substringBefore(String str, String separator) {
		return org.apache.commons.lang3.StringUtils.substringBefore(str, separator);
	}

	public static String substringBeforeLast(String str, String separator) {
		return org.apache.commons.lang3.StringUtils.substringBeforeLast(str, separator);
	}

	public static String substringAfter(String str, String separator) {
		return org.apache.commons.lang3.StringUtils.substringAfter(str, separator);
	}

	public static String substringAfterLast(String str, String separator) {
		return org.apache.commons.lang3.StringUtils.substringAfterLast(str, separator);
	}

	public static String removeEnd(String str, String toRemove) {
		return org.apache.commons.lang3.StringUtils.removeEnd(str, toRemove);
	}

	public static String sanitizeHTML(String html, List<SanitizerPolicy> policyParams) {
		PolicyFactory policyFactory = null;

		for (SanitizerPolicy param : policyParams) {
			policyFactory = (policyFactory == null) ? SANITIZER_POLICIES.get(param.name()) : policyFactory.and(SANITIZER_POLICIES.get(param.name()));
		}

		return sanitizeHTML(html, policyFactory);
	}

	public static String sanitizeHTML(String html, PolicyFactory policyFactory) {
		return policyFactory.sanitize(html);
	}

	public static String stringSimplify(String value) {
		String normalized = Normalizer.normalize(value, Normalizer.Form.NFD);
		return normalized.replaceAll("\\p{M}", ""); // removes all characters in Unicode Mark category
	}

	public static Boolean isStringSimplified(String value) {
		return Normalizer.isNormalized(value, Normalizer.Form.NFD);
	}
}
