# METADATA
# scope: package
# title: Strong password policy
# description: Bruteforce is quite common. Ensure passwords are very strong.
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: Security
#  severity: HIGH
#  rulename: StrongPasswordPolicy
#  priority: 5
#  rulenumber: 001_0004
#  remediation: Ensure minimum password length of at least 8 characters and must use all character classes.
#  input: Security$ProjectSecurity.yaml
package app.mendix.project_settings.strong_password
import rego.v1
annotation := rego.metadata.chain()[1].annotations

default allow := false
allow if count(errors) == 0

min_password_length := 8

errors contains error if {
    my_password_length := input.PasswordPolicySettings.MinimumLength
    my_password_length < min_password_length
    error := sprintf("[%v, %v, %v] Password length of %v is not enough. It must be at least %v",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            my_password_length,
            min_password_length,
        ]
    )
}

errors contains error if {
    input.PasswordPolicySettings.RequireDigit == false
    error := sprintf("[%v, %v, %v] %v",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            "Password must require digits",
        ]
    )
}

errors contains error if {
    input.PasswordPolicySettings.RequireMixedCase == false
    input.PasswordPolicySettings.RequireSymbol == false
    error := sprintf("[%v, %v, %v] %v",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            "Password must require mixed case characters",
        ]
    )
}

errors contains error if {
    input.PasswordPolicySettings.RequireSymbol == false
    error := sprintf("[%v, %v, %v] %v",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            "Password must require symbols",
        ]
    )
}