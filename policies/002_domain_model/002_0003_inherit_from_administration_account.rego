# METADATA
# scope: package
# title: Inherit from Administration.Account
# description: There is no need to inherit from administration.account. Administration.account may simply be extended, this is not a system module. Avoid unnecessary inheritance as this has a negative effect on performance.
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: Performance
#  rulename: AvioidInheritanceFromAdministrationAccount
#  severity: MEDIUM
#  rulenumber: 002_0003
#  remediation: Inherit from system.user instead or adapt Administration.Account so it fits your needs.
#  input: "*/DomainModels$DomainModel.yaml"
package app.mendix.domain_model.inherit_from_administration_account
import rego.v1
annotation := rego.metadata.chain()[1].annotations

default allow := false
allow if count(errors) == 0

errors contains error if {
    some i
    input.Entities[i].MaybeGeneralization.Generalization == "Administration.Account"
    error := sprintf("[%v, %v, %v] Entity %v has generaralization of %v",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            [input.Entities[i].Name],
            "Administration.Account",
        ])
}