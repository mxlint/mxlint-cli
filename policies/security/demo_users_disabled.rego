# input: Security$ProjectSecurity.yaml
package app.mendix.security

import rego.v1

default demo_users_disabled := false

demo_users_disabled if {
    input.EnableDemoUsers == false
}