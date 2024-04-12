package app.mendix.pages.inline_style_property_used
import rego.v1


# Test data
form_simple = {
	"$Type": "Forms$Page",
	"Name": "Page1",
	"Appearance": {
		"$Type": "Forms$Appearance",
		"Class": "",
		"DesignProperties": null,
		"DynamicClasses": "",
		"Style": "",
	},
}
form_simple_negative = {
	"$Type": "Forms$Page",
	"Name": "Page1",
	"Appearance": {
		"$Type": "Forms$Appearance",
		"Class": "",
		"DesignProperties": null,
		"DynamicClasses": "",
		"Style": "color: red;",
	},
}

form_nested = {
	"Name": "Page1",
	"FormCall": {
		"Arguments": [
			{
				"Widgets": [
					{
						"$Type": "Forms$LayoutGrid",
						"Name": "layoutGrid2",
						"Rows": [
							{
								"$Type": "Forms$LayoutGridRow",
								"Columns": [
									{
										"$Type": "Forms$LayoutGridColumn",
										"Appearance": {
											"$Type": "Forms$Appearance",
											"Class": "",
											"DesignProperties": null,
											"DynamicClasses": "",
											"Style": "",
										}
									},
								],
							},
						],
					},
				],
			},
		],
	},
}

form_nested_negative = {
	"Name": "Page1",
	"FormCall": {
		"Arguments": [
			{
				"Widgets": [
					{
						"$Type": "Forms$LayoutGrid",
						"Name": "layoutGrid2",
						"Rows": [
							{
								"$Type": "Forms$LayoutGridRow",
								"Columns": [
									{
										"$Type": "Forms$LayoutGridColumn",
										"Appearance": {
											"$Type": "Forms$Appearance",
											"Class": "",
											"DesignProperties": null,
											"DynamicClasses": "",
											"Style": "color: orange;",
										}
									},
								],
							},
						],
					},
				],
			},
		],
	},
}



# Test cases
test_simple if {
	allow with input as form_simple
}

test_simple_negative if {
	not allow with input as form_simple_negative
}

test_nested if {
	allow with input as form_nested
}

test_nested_negative if {
	not allow with input as form_nested_negative
}