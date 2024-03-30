package app.mendix.project_settings.strong_password
import rego.v1

# Test cases
test_allow if {
	allow with input as {
		"PasswordPolicySettings": {
			"MinimumLength": 9,
			"RequireDigit": true,
			"RequireSymbol": true,
			"RequireMixedCase": true,
		}
	}
}

test_no_allow_password_length if {
	not allow with input as {
		"PasswordPolicySettings": {
			"MinimumLength": 3,
			"RequireDigit": true,
			"RequireSymbol": true,
			"RequireMixedCase": true,
		}
	}
}

test_no_allow_simple if {
	not allow with input as {
		"PasswordPolicySettings": {
			"MinimumLength": 3,
			"RequireDigit": false,
			"RequireSymbol": true,
			"RequireMixedCase": false,
		}
	}
}