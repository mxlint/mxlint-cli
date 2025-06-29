# METADATA
# scope: package
# title: Ensure security rules are active
# description: Any serious app needs entity access security configured
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: Security
#  rulename: SecurityChecks
#  severity: HIGH
#  rulenumber: 001_0003
#  remediation: Set Security check to production in Project Security
#  input: .*Security\$ProjectSecurity\.yaml
package app.mendix.project_settings.security_checks
import rego.v1
annotation := rego.metadata.chain()[1].annotations

default allow := false
allow if count(errors) == 0

errors contains error if {
    input.CheckSecurity == false
    error := sprintf("[%v, %v, %v] %v",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            "Security check is not enabled in Project Security",
        ]
    )
}

errors contains error if {
    input.SecurityLevel != "CheckEverything" 
    error := sprintf("[%v, %v, %v] %v",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            "Security check is not set to Production in Project Security",
        ]
    )
}