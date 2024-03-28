# METADATA
# title: Business apps must require login
# description: No anonymous means every user must have valid login session or credentials
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: security
#  rulename: AnonymousDisabled
#  priority: 4
#  rulenumber: 001-0001
#  remediation: Disable anonymous/guest access in Project Security
#  input: Security$ProjectSecurity.yaml

package app.mendix.projectsettings

import rego.v1

default anonymous_disabled := false

anonymous_disabled if {
    input.EnableGuestAccess == false
}