# METADATA
# title: Ensure security rules are active
# description: Any serious app needs entity access security configured
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: security
#  rulename: SecurityEnabled
#  priority: 5
#  rulenumber: 001-0003
#  remediation: Turn on Security check in Project Security
#  input: Security$ProjectSecurity.yaml

package app.mendix.projectsettings

import rego.v1

default security_enabled := false

security_enabled if {
    input.CheckSecurity == true
    input.SecurityLevel == "CheckEverything"
}