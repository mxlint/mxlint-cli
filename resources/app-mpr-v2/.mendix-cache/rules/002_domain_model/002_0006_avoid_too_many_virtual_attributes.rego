# METADATA
# scope: package
# title: Too many Microflow attributes (virtual attributes) inside of an entity
# description: Too many Microflow attributes (virtual attributes) inside of an entity will cause performance issues.
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: Performance
#  rulename: AvoidTooManyVirtualAttributes
#  severity: MEDIUM
#  rulenumber: 002_0006
#  remediation: Optimize the number of virtual attributes inside of an entity. Reduce to 10 or less.
#  input: "*/DomainModels$DomainModel.yaml"
package app.mendix.domain_model.avoid_too_many_virtual_attributes

import rego.v1
annotation := rego.metadata.chain()[1].annotations

default allow := false
allow if count(errors) == 0

max_virtual_attributes := 10

errors contains error if {
    not input.Entities == null
    
    entity := input.Entities[_]
    entity_name := entity.Name
    attr_count := count([attr | attr := entity.Attributes[_]; attr.Value["$Type"] == "DomainModels$CalculatedValue"])
    attr_count > max_virtual_attributes
    error := sprintf("[%v, %v, %v] There are %v Virtual Attributes in entity %v which is more than %v",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            attr_count,
            entity_name,
            max_virtual_attributes
        ]
    )
}