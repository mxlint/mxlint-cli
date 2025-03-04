# METADATA
# scope: package
# title: No more that 35 attributes in an entity
# description: The bigger the entities, the slower your application will become when handling the data. This is because Mendix is using SELECT * queries a lot and will retrieve a lot of unnecessary data.
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: Maintainability
#  rulename: NumberOfAttributes
#  severity: MEDIUM
#  rulenumber: 002_0002
#  remediation: Normalize your datamodel. Split your object into multiple objects. If the attributes really belong to each other in a one-to-one relation, just draw a one-to-one relation between the objects.
#  input: "*/DomainModels$DomainModel.yaml"
package app.mendix.domain_model.number_of_attributes
import rego.v1
annotation := rego.metadata.chain()[1].annotations

default allow := false
allow if count(errors) == 0

max_attributes := 35

errors contains error if {
    entity := input.Entities[_]
    not entity.Attributes == null
    count_attributes := count(entity.Attributes)
    count_attributes > max_attributes
    error := sprintf("[%v, %v, %v] Entity %v has %v attributes which is more than %v",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            entity.Name,
            count_attributes,
            max_attributes
        ]
    )
}