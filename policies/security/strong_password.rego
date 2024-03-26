# input: Security$ProjectSecurity.yaml
package app.mendix.security

import rego.v1

default strong_password := false

strong_password if {
    input.PasswordPolicySettings.MinimumLength >= 8
    input.PasswordPolicySettings.RequireDigit == true
    input.PasswordPolicySettings.RequireMixedCase == true
    input.PasswordPolicySettings.RequireSymbol == true
}