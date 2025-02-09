# METADATA
# scope: package
# title: Avoid using system storage objects directly
# description: Always inherit for filedocuments and images. Never implement direct assocations to the System Domain Model, because of limits on the configuration of security.
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: Security
#  rulename: AvoidSystemEntityAssociation
#  severity: HIGH
#  rulenumber: 002_0005
#  remediation: Remove direct associations with the System Domain Model. Use inheritance instead (i.e. Generalization in the entity properties).
#  input: "*/DomainModels$DomainModel.yaml"
package app.mendix.domain_model.avoid_system_entity_association
import rego.v1
annotation := rego.metadata.chain()[1].annotations

default allow := false
allow if count(errors) == 0

errors contains error if {
    some i
    startswith(input.CrossAssociations[i].Child, "System.")
    error := sprintf("[%v, %v, %v] Entity association %v refers to a System entity with limited security configuration.",
    [
        annotation.custom.severity,
        annotation.custom.category,
        annotation.custom.rulenumber,
        [input.CrossAssociations[i].Name],
    ])
}