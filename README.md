# Mendix Model Exporter

Mendix models are stored in a binary file with `.mpr` extension. This project aims to export Mendix models to a human readable format, such as Yaml. This enables developers to use traditional code analysis tools on Mendix models. Think of quality checks like linting, code formatting, etc.

![Mendix Model Exporter](./resources/model-new-entity.png)

See each Mendix document/object as a separate file in the output directory. And see the differences between versions in a version control system. Here we changed the `Documentation` of an entity and added a new `Entity` with one `Attribute`.

## Usage

### Download

First download the exporter program. 

- Download the latest release from the [releases page](https://github.com/cinaq/mendix-model-exporter/releases)
- Run the executable with the following command line arguments:
  - `--input` or `-i` to specify the input file or directory. Default is `./` which means the current directory. It will look for `.mpr` files in the current directory.
  - `--output` or `-o` to specify the output directory. Default is `modelsource`

## pre-commit hook setup

> As of this writing, Mendix Studio Pro does not support Git hooks. You can use the following workaround to automatically export your Mendix model to Yaml before each commit. Make sure you have git-bash installed on your system. Download it from [here](https://git-scm.com/download/win).

After you open your `git` project in Mendix Studio Pro, navigate to the root of the project. Create a new file named `.git/hooks/pre-commit` and add the following content:

```bash
#!/bin/sh

# Program name and download URL
PROGRAM_NAME="mendix-model-exporter"
## might need to change the version. See for latest version at https://github.com/cinaq/mendix-model-exporter/releases
VERSION="v1.0.0"

# Path to the program in the hooks directory
PROGRAM_PATH="$(dirname $0)/$PROGRAM_NAME"

# Check if the program exists, download it if it does not
if [ ! -f "$PROGRAM_PATH" ]; then
  echo "Program not found, downloading..."
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m | tr '[:upper:]' '[:lower:]')"
  if [ "$OS" = "windows" ]; then
    EXT=".exe"
  else
    EXT=""
  fi
  if [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
  fi
  DOWNLOAD_URL="https://github.com/cinaq/mendix-model-exporter/releases/download/$VERSION/mendix-model-exporter-$VERSION-$OS-$ARCH$EXT"
  curl -L -sf "$DOWNLOAD_URL" -o "$PROGRAM_PATH"
  chmod +x "$PROGRAM_PATH"
fi

# Execute the program
"$PROGRAM_PATH"

# Check program execution result
if [ $? -ne 0 ]; then
  echo "Program failed, aborting commit."
  exit 1
fi

# Automatically stage changes made by your program
git add modelsource

# Exit with 0 to continue the commit process
exit 0
```

Set the executable bit on the file:

```bash
chmod +x .git/hooks/pre-commit
```

Now whenever you commit using git-bash, the Mendix model will be exported to Yaml before the commit. The changes will be staged automatically.

## Contribute

Create a PR with your changes. We will review and merge it.

## Features

- Export Mendix model to Yaml
- Incremental changes
- Human readable output

## TODO

- [x] Export Mendix model to Yaml
- [ ] Expand test coverage
- [ ] Support incremental changes
- [ ] Improve performance for large models
- [ ] Improve error handling
- [ ] Improve output human readability