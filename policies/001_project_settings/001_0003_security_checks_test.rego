package app.mendix.project_settings.security_checks
import rego.v1

# Test cases
test_allow if {
	allow with input as {
		"CheckSecurity": true,
		"SecurityLevel": "CheckEverything",
	}
}
test_no_allow_1 if {
	not allow with input as {
		"CheckSecurity": false,
		"SecurityLevel": "CheckEverything",
	}
}
test_no_allow_2 if {
	not allow with input as {
		"CheckSecurity": true,
		"SecurityLevel": "unknown",
	}
}