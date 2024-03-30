package app.mendix.project_settings.anonymous_disabled
import rego.v1

# Test cases
test_allow if {
	allow with input as {"EnableGuestAccess": false}
}
test_no_allow if {
	not allow with input as {"EnableGuestAccess": true}
}