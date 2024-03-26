# input: Security$ProjectSecurity.yaml
package app.mendix.security

import rego.v1

default security_enabled := false

security_enabled if {
    input.CheckSecurity == true
    input.SecurityLevel == "CheckEverything"
}