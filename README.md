# Mendix CLI

A set of Command line interface tools for Mendix developers, CICD engineers and platform engineers.

> This project is in early development stage. Please use with caution. We are looking for contributors to help us improve the tools. Please create a PR with your changes. We believe in open ecosystem and open source. We are looking forward to your contributions. These can be documentation improvements, bug fixes, new features, etc.

## export-model

Mendix models are stored in a binary file with `.mpr` extension. This project exports Mendix model to a human readable format, such as Yaml. This enables developers to use traditional code analysis tools on Mendix models. Think of quality checks like linting, code formatting, etc.

![Mendix Model Exporter](./resources/model-new-entity.png)

See each Mendix document/object as a separate file in the output directory. And see the differences between versions in a version control system. Here we changed the `Documentation` of an entity and added a new `Entity` with one `Attribute`.

### pre-commit hook setup

> As of this writing, Mendix Studio Pro does not support Git hooks. You can use the following workaround to automatically export your Mendix model to Yaml before each commit. Make sure you have git-bash installed on your system. It is already present if you use Mendix 10. If you don't have it, download from [here](https://git-scm.com/download/win).

Open git-bash inside of your project directory (use `cd Mendix/project-name` if needed) and run the following commands:

```bash
curl https://github.com/cinaq/mendix-cli/raw/main/pre-commit -o .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

# try it out
.git/hooks/pre-commit
```

Now whenever you commit using git-bash, the Mendix model will be exported to Yaml before the commit. The changes will be included automatically.

#### Pipeline integration

If you do not want to export the model to Yaml on your local machine, you can do it in your pipeline. Here's a high-level example:

```bash
$ ./bin/mendix-cli-darwin-arm64 export-model -i resources/full-app-v1.mpr
INFO[0000] Exporting resources/full-app-v1.mpr to modelsource
INFO[0000] Completed resources/full-app-v1.mpr

$ ./bin/mendix-cli-darwin-arm64 lint
## policies/001_project_settings/001_0001_anonymous_disabled.rego
PASS (0.00148s) modelsource/Security$ProjectSecurity.yaml

## policies/001_project_settings/001_0002_demo_users_disabled.rego
SKIP (0.00000s) modelsource/Security$ProjectSecurity.yaml

## policies/001_project_settings/001_0003_security_checks.rego
SKIP (0.00000s) modelsource/Security$ProjectSecurity.yaml

## policies/001_project_settings/001_0004_strong_password.rego
SKIP (0.00000s) modelsource/Security$ProjectSecurity.yaml

## policies/002_domain_model/002_0001_number_of_entities.rego
PASS (0.00190s) modelsource/Administration/DomainModels$DomainModel.yaml
PASS (0.00156s) modelsource/Atlas_UI_Resources/DomainModels$DomainModel.yaml
PASS (0.00240s) modelsource/MyFirstModule/DomainModels$DomainModel.yaml

## policies/002_domain_model/002_0002_number_of_attributes.rego
PASS (0.00161s) modelsource/Administration/DomainModels$DomainModel.yaml
PASS (0.00113s) modelsource/Atlas_UI_Resources/DomainModels$DomainModel.yaml
PASS (0.00158s) modelsource/MyFirstModule/DomainModels$DomainModel.yaml
```

## lint

![Mendix Lint report](./resources/lint-xunit-report.png)
Lint Mendix Yaml files. This tool checks for common mistakes and enforces best practices. It uses OPA as policy engine. Therefore policies must be written in the powerful Rego language. Please refer to [Rego language reference](https://www.openpolicyagent.org/docs/latest/policy-reference/) for more information on the syntax and semantics.

### Features

- Export Mendix model to Yaml
- Lint Mendix Yaml files for common mistakes and enforces best practices
- Incremental changes
- Human readable output

## TODO

- [x] Export Mendix model to Yaml
- [x] Improve output human readability
- [x] Linting for Mendix Yaml files
- [x] Create policies for linting
- [ ] Expand test coverage
- [ ] Support incremental changes
- [ ] Improve performance for large models
- [ ] Improve error handling
- [ ] Transform flows (activities, decisions, etc.) to pseudo code

## Contribute

Create a PR with your changes. We will review and merge it.

Rego files must follow the [style guide](https://github.com/StyraInc/rego-style-guide/blob/main/style-guide.md)

Make sure to run the tests before creating a PR:

```bash
make test
```

## License

Code is licensed under the [MIT License](./LICENSE). [CINAQ](https://cinaq.com) is a registered trademark of CINAQ B.V.. Mendix is a registered trademark of Mendix B.V. All other trademarks are the property of their respective owners.