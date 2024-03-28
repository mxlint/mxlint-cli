# METADATA
# title: Strong password policy
# description: Bruteforce is quite common. Ensure passwords are very strong.
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: security
#  rulename: StrongPasswordPolicy
#  priority: 5
#  rulenumber: 001-0004
#  remediation: Ensure minimum password length of at least 8 characters and must use all character classes.
#  input: Security$ProjectSecurity.yaml

package app.mendix.projectsettings

import rego.v1

default strong_password_policy := false

strong_password_policy if {
    input.PasswordPolicySettings.MinimumLength >= 8
    input.PasswordPolicySettings.RequireDigit == true
    input.PasswordPolicySettings.RequireMixedCase == true
    input.PasswordPolicySettings.RequireSymbol == true
}