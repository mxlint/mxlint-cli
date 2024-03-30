package app.mendix.project_settings.demo_users_disabled
import rego.v1

# Test cases
test_allow if {
	allow with input as {"EnableDemoUsers": false}
}
test_no_allow if {
	not allow with input as {"EnableDemoUsers": true}
}