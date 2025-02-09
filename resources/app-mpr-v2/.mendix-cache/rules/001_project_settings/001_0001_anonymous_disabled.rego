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