# METADATA
# scope: package
# title: No more than 15 persistent entities within one domain model
# description: The bigger the domain models, the harder they will be to maintain. It adds complexity to your security model as well. The smaller the modules, the easier to reuse.
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: Maintainability
#  rulename: NumberOfEntities
#  severity: MEDIUM
#  rulenumber: 002_0001
#  remediation: Split domain model into multiple modules.
#  input: "*/DomainModels$DomainModel.yaml"
package app.mendix.domain_model.number_of_entities
import rego.v1
annotation := rego.metadata.chain()[1].annotations

default allow := false
allow if count(errors) == 0

max_entities := 15

errors contains error if {
    not input.Entities == null
    count_entities := count(input.Entities)
    count_entities > max_entities
    error := sprintf("[%v, %v, %v] There are %v entities which is more than %v",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            count_entities,
            max_entities
        ]
    )
}