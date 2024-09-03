package app.mendix.domain_model.avoid_using_validation_rules

import rego.v1


# Test data

positive := {
  "Entities": [
    {
      "ValidationRules": [],
      "Name": "Bike"
    }
  ]
}

negative := {
  "Entities": [
    {
      "ValidationRules": [
        {
          "$Type": "DomainModels$ValidationRule",
          "Attribute": "MyFirstModule.Bike.Name",
          "Message": {
            "$Type": "Texts$Text",
            "Items": [
              {
                "$Type": "Texts$Translation",
                "LanguageCode": "en_US",
                "Text": "Not a good name"
              }
            ]
          },
          "RuleInfo": {
            "$Type": "DomainModels$EqualsToRuleInfo",
            "EqualsToAttribute": "",
            "UseValue": true,
            "Value": "admin"
          }
        }
      ],
      "Name": "Bike"
    }
  ]
}

# Test cases

test_positive if {
	allow with input as positive
}

test_negative if {
	not allow with input as negative
}
