# METADATA
# scope: package
# title: Inherit from non System module is discouraged
# description: Inheritance, except from system module, is strongly discouraged because of the negative performance side effects.
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: Performance
#  rulename: AvoidInheritanceFromNonSystem
#  severity: MEDIUM
#  rulenumber: 002_0004
#  remediation: Instead of inheritance, just use separate objects which are associated to the main object. As an alternative, you can add the child’s attributes to the super entity and add an ObjectType enumeration.
#  input: "*/DomainModels$DomainModel.yaml"
package app.mendix.domain_model.inherit_from_non_system
import rego.v1
annotation := rego.metadata.chain()[1].annotations

default allow := false
allow if count(errors) == 0

errors contains error if {
    some i
    not startswith(input.Entities[i].MaybeGeneralization.Generalization, "System.")
    error := sprintf("[%v, %v, %v] Entity %v has generaralization of non-System",
    [
        annotation.custom.severity,
        annotation.custom.category,
        annotation.custom.rulenumber,
        [input.Entities[i].Name],
    ])
}