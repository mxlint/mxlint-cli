# METADATA
# title: Business apps should not have demo users
# description: No demo users 
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: security
#  rulename: DemoUsersDisabled
#  priority: 5
#  skip: FIXME
#  rulenumber: 001-0002
#  remediation: Disable demo users in Project Security
#  input: Security$ProjectSecurity.yaml
package app.mendix.security

import rego.v1

default demo_users_disabled := false

demo_users_disabled if {
    input.EnableDemoUsers == false
}