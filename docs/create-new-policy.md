## Create new policy

### Rego Introduction

Policies are expressed with the help of the powerful OPA Rego language. Rego is a declarative language that is purpose-built for expressing policies over complex hierarchical data structures. Rego is designed to be easy to read and write, even for non-programmers. Rego is a safe language that is decidable and has a small trusted computing base. Rego is also designed to be easy to integrate with other systems.


### Policy rule

To create a new policy, you need to create a new Rego file in the `policies` directory. The file name should be in the format `XXX_YYY.rego` where `XXX` is the policy number and `YYY` is the policy name. For example, `001_0001_anonymous_disabled.rego`.

The policy file should contain the following structure:

```rego
# METADATA
# scope: package
# title: Business apps must always require login
# description: No anonymous means every user must have valid login session or credentials
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: Security
#  rulename: AnonymousDisabled
#  severity: HIGH
#  rulenumber: 001_0001
#  remediation: Disable anonymous/guest access in Project Security
#  input: Security$ProjectSecurity.yaml
package app.mendix.project_settings.anonymous_disabled
import rego.v1
annotation := rego.metadata.chain()[1].annotations

default allow := false
allow if count(errors) == 0

errors contains error if {
    input.EnableGuestAccess == true
    error := sprintf("[%v, %v, %v] %v",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            annotation.title,
        ]
    )
}
```

- `METADATA` provide information about the policy. 
- `package` statement is used to define the policy package. 
- `allow` statement is used to define the conditions under which the policy is allowed. 
- `errors` statement is used to define the errors that are returned if the policy is not allowed.
- `input` states which files are used as input for the policy. This can be a single file or an expression like `*/DomainModels$DomainModel.yaml` to match multiple files.

## Policy testing

The best way to create a new policy is to copy an existing policy and modify it to suit your needs. There is also an accompanying test file for each policy that you can use to test your policy. The test file should be in the same directory as the policy file and should be named `XXX_YYY_test.rego`. For example, `001_0001_anonymous_disabled_test.rego`.

```rego
package app.mendix.project_settings.anonymous_disabled
import rego.v1

# Test cases
test_allow if {
	allow with input as {"EnableGuestAccess": false}
}
test_no_allow if {
	not allow with input as {"EnableGuestAccess": true}
}
```

To test your policy, run the following command:

```bash
$ ./run-policy-tests.sh              
policies/001_project_settings/001_0001_anonymous_disabled_test.rego:
data.app.mendix.project_settings.anonymous_disabled.test_allow: PASS (3.031209ms)
data.app.mendix.project_settings.anonymous_disabled.test_no_allow: PASS (413.375µs)

policies/001_project_settings/001_0002_demo_users_disabled_test.rego:
data.app.mendix.project_settings.demo_users_disabled.test_allow: PASS (105.541µs)
data.app.mendix.project_settings.demo_users_disabled.test_no_allow: PASS (200.5µs)

policies/001_project_settings/001_0003_security_checks_test.rego:
data.app.mendix.project_settings.security_checks.test_allow: PASS (111.584µs)
data.app.mendix.project_settings.security_checks.test_no_allow_1: PASS (842.667µs)
data.app.mendix.project_settings.security_checks.test_no_allow_2: PASS (206.458µs)

policies/001_project_settings/001_0004_strong_password_test.rego:
data.app.mendix.project_settings.strong_password.test_allow: PASS (148.792µs)
data.app.mendix.project_settings.strong_password.test_no_allow_password_length: PASS (538.959µs)
data.app.mendix.project_settings.strong_password.test_no_allow_simple: PASS (286.916µs)

policies/002_domain_model/002_0001_number_of_entities_test.rego:
data.app.mendix.domain_model.number_of_entities.test_no_entities: PASS (134µs)
data.app.mendix.domain_model.number_of_entities.test_1_entity: PASS (194.666µs)
data.app.mendix.domain_model.number_of_entities.test_2_entities: PASS (187.334µs)
data.app.mendix.domain_model.number_of_entities.test_20_entities: PASS (1.375709ms)

policies/002_domain_model/002_0002_number_of_attributes_test.rego:
data.app.mendix.domain_model.number_of_attributes.test_no_entities: PASS (263.5µs)
data.app.mendix.domain_model.number_of_attributes.test_1_entity_1_attribute: PASS (519.416µs)
data.app.mendix.domain_model.number_of_attributes.test_2_entities: PASS (303.458µs)
data.app.mendix.domain_model.number_of_attributes.test_3_entities_1_empty: PASS (342.958µs)
data.app.mendix.domain_model.number_of_attributes.test_1_entity_40_attributes_not_allowed: PASS (1.294166ms)
data.app.mendix.domain_model.number_of_attributes.test_2_entity_40_attributes_1_empty_not_allowed: PASS (2.156042ms)
--------------------------------------------------------------------------------
PASS: 20/20

```

The test rego files contain examples so that it's easy to validate your policy actually works for different scenarios with purpose-crafted input data. The test script will run all the test files in the `policies` directory and output the results.

Example input could be inspected in the `modelsource` directory. The `modelsource` directory contains the exported Mendix model in Yaml format. The `modelsource` directory is created when you run the `export-model` command.


### Help

We understand Rego is not the easiest language to use. However, it is the perfect match due to its expressiveness. If you need help creating a new policy, please reach out to us at support@cinaq.com.