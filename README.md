# MxLint CLI

A set of Command line interface tools for Mendix developers, CICD engineers and platform engineers.

> This project is in early development stage. Please use with caution. We are looking for contributors to help us improve the tools. Please create a PR with your changes. We believe in open ecosystem and open source. We are looking forward to your contributions. These can be documentation improvements, bug fixes, new features, etc.

## Mendix Studio Pro extension

The quickest way to try out MxLint is to use it as Mendix Studio pro extension. Follow the instructions at [mxlint-extension](https://github.com/mxlint/mxlint-extension)

## Basic Usage

mxlint-cli is a set of tools to help you with your Mendix projects. As such you can use it in many ways. To give you a feeling what it does. Try the following example:

### Prerequisites

- Mendix Project source
- Operating system: Linux, MacOS, Windows
- Download your platform specific binary from the [releases page](https://github.com/mxlint/mxlint-cli/releases)
- Download the policies from the [releases page](https://github.com/mxlint/mxlint-rules/releases) and extract them to a directory

### Export Mendix model to Yaml

- copy `mxlint-cli` to your project directory
- Open a terminal and navigate to your project directory; ideally use git-bash on Windows or Terminal on MacOS/Linux
- run `./mxlint-cli export-model`

You will see a new directory `modelsource` with the exported Mendix model in Yaml format

It's advisable to add the `mxlint-cli` file to your `.gitignore` file. This way you don't accidentally commit it to your repository.

### Lint Mendix Yaml files

- copy `policies` directory to your project directory
- run `./mxlint-cli lint --xunit-report=report.xml`

You will see a summary of the policy evaluations in the terminal and a report in the `report.xml` file. The report is in xUnit format. You can use it in your CI/CD pipeline.

Do you want to create your own policies? Please refer to our guide [Create new policy](./docs/create-new-policy.md)

## export-model

Mendix models are stored in a binary file with `.mpr` extension. This project exports Mendix model to a human readable format, such as Yaml. This enables developers to use traditional code analysis tools on Mendix models. Think of quality checks like linting, code formatting, etc.

![Mendix Model Exporter](./resources/model-new-entity.png)

See each Mendix document/object as a separate file in the output directory. And see the differences between versions in a version control system. Here we changed the `Documentation` of an entity and added a new `Entity` with one `Attribute`.

#### Pipeline integration

If you do not want to export the model to Yaml on your local machine, you can do it in your pipeline. Here's a high-level example:

```bash
$ ./bin/mxlint-cli-darwin-arm64 export-model -i resources/full-app-v1.mpr
INFO[0000] Exporting resources/full-app-v1.mpr to modelsource
INFO[0000] Completed resources/full-app-v1.mpr

$ ./bin/mxlint-cli-darwin-arm64 lint
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

### NOQA (Ignore document)

A specific document can be marked as "Skipped" if you have a line in the `documentation` field that starts with either `#noqa` or `# noqa` followed by an optional message (Case in-sensitive). This message will be included as "Skipped" reason in linting results.

## watch

Watch for changes in the model and lint the changes.

```
./bin/mxlint-darwin-arm64 watch --input resources/app/ --rules resources/rules
FILE "triggered event" CREATE [-]
INFO[0000] Watching for changes in /Users/xcheng/private/git/mxlint-cli/resources/app 
INFO[0000] Output directory: modelsource                
INFO[0000] Rules directory: resources/rules             
INFO[0000] Mode: basic                                  
INFO[0000] Exporting resources/app/App.mpr to modelsource 
INFO[0000] Transforming microflow UpdateUserHelper      
INFO[0000] Transforming microflow AssertTrue            
INFO[0000] Transforming microflow CreateUserIfNotExists 
INFO[0000] Transforming microflow AssertTrue_2          
INFO[0000] Transforming microflow ChangeMyPassword      
INFO[0000] Transforming microflow ShowMyPasswordForm    
INFO[0000] Transforming microflow ManageMyAccount       
INFO[0000] Loop detected; not traversing                
INFO[0000] Transforming microflow NewAccount            
INFO[0000] Transforming microflow ChangePassword        
INFO[0000] Transforming microflow NewWebServiceAccount  
INFO[0000] Transforming microflow ShowPasswordForm      
INFO[0000] Transforming microflow SaveNewAccount        
INFO[0000] Transforming microflow MicroflowSplit        
INFO[0000] Transforming microflow MicroflowSimple       
INFO[0000] Transforming microflow MicroflowComplexSplit 
INFO[0000] Loop detected; not traversing                
INFO[0000] Transforming microflow MicroflowLoopNested   
INFO[0000] Loop detected; not traversing                
INFO[0000] Loop detected; not traversing                
INFO[0000] Loop detected; not traversing                
INFO[0000] Loop detected; not traversing                
INFO[0000] Transforming microflow MicroflowSplitThenMerge 
INFO[0000] Loop detected; not traversing                
INFO[0000] Transforming microflow MicroflowLoop         
INFO[0000] Loop detected; not traversing                
INFO[0000] Loop detected; not traversing                
INFO[0000] Transforming microflow MyFirstLogic          
INFO[0000] Transforming microflow MicroflowForLoop      
INFO[0000] Transforming microflow VA_Age                
INFO[0000] Found 361 documents                          
INFO[0000] Completed resources/app/App.mpr              
## resources/rules/001_0003_security_checks.rego
FAIL (0.00171s) modelsource/Security$ProjectSecurity.yaml

WARN[0000] Rule resources/rules/001_0003_security_checks.rego: 1 failures 
WARN[0000]   Document modelsource/Security$ProjectSecurity.yaml: [HIGH, Security, 4099] Security check is not enabled in Project Security 
WARN[0000] Lint failed: 1 failures 
```

## test-rules

Rules can be written in both `Rego` and `JavaScript` format. To speed up rule development we have implemented `test-rules` subcommand that can quickly evaluate your rule against known test scenarios. The test cases are written in `yaml` format. 

```
$ ./bin/mxlint-darwin-arm64 test-rules -r resources/rules
INFO[0000] >> resources/rules/001_0002_demo_users_disabled.js 
INFO[0000] PASS  allow
INFO[0000] PASS  no_allow
INFO[0000] >> resources/rules/001_0003_security_checks.rego 
INFO[0000] PASS  allow
INFO[0000] PASS  no_allow_1
INFO[0000] PASS  no_allow_2
```

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
- [x] Output linting results in xUnit format
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

This project is an initiative of CINAQ. See [LICENSE](./LICENSE). [CINAQ](https://cinaq.com) is a registered trademark of CINAQ B.V.. Mendix is a registered trademark of Mendix B.V. All other trademarks are the property of their respective owners.
