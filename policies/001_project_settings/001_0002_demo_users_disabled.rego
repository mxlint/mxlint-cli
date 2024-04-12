# METADATA
# scope: package
# title: Business apps should disable demo users
# description: No demo users 
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: Security
#  rulename: DemoUsersDisabled
#  severity: HIGH
#  rulenumber: 001_0002
#  remediation: Disable demo users in Project Security
#  input: Security$ProjectSecurity.yaml
package app.mendix.project_settings.demo_users_disabled
import rego.v1
annotation := rego.metadata.chain()[1].annotations

default allow := false
allow if count(errors) == 0

errors contains error if {
    input.EnableDemoUsers == true
    error := sprintf("[%v, %v, %v] %v",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            annotation.title,
        ]
    )
}