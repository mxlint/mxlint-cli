# METADATA
# scope: package
# title: Avoid using validation rules for domain model.
# description: Validation rules on domain model level will give the users unexpected errors.
# authors:
# - Viktor Berlov <viktor@cinaq.com>
# custom:
#  category: Maintainability
#  rulename: AvoidUsingValidationRules
#  severity: MEDIUM
#  rulenumber: 002_0007
#  remediation: Remove datamodel validation rules.
#  input: "*/DomainModels$DomainModel.yaml"
package app.mendix.domain_model.avoid_using_validation_rules

import rego.v1
annotation := rego.metadata.chain()[1].annotations

default allow := false

allow if count(errors) == 0

errors contains error if {
    entity := input.Entities[_]
    entity_name := entity.Name
    rules_count := count([rule | rule := entity.ValidationRules[_]])
    rules_count > 0
    
    error := sprintf("[%v, %v, %v] Validation rules %v in entity %v",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            rules_count,
            entity_name
        ]
    )
}
