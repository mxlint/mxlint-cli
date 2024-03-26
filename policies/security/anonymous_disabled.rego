# input: Security$ProjectSecurity.yaml
package app.mendix.security

import rego.v1

default anonymous_disabled := false

anonymous_disabled if {
    input.EnableGuestAccess == false
}