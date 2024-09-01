package app.mendix.domain_model.avoid_too_many_virtual_attributes

import rego.v1


# Test data


attr_0 := {
          "$Type": "DomainModels$Attribute",
          "Name": "VA_age",
          "Value": {
            "$Type": "DomainModels$CalculatedValue"
          }
}


twenty := numbers.range(1, 20)
attr_20 = [ 
    { "Name": attr_0.Name, "Value": attr_0.Value }  | n := twenty[_]
]

positive := {
  "Entities": [
    {
      "$Type": "DomainModels$EntityImpl",
      "Attributes": [
        {
          "$Type": "DomainModels$Attribute",
          "Name": "VA_age",
          "Value": {
            "$Type": "DomainModels$CalculatedValue"
          }
        },
        {
          "$Type": "DomainModels$Attribute",
          "Name": "Year",
          "Value": {
            "$Type": "DomainModels$StoredValue"
          }
        }
      ],
      "Name": "Bike"
    }
  ]
}

negative := {
  "Entities": [
    {
      "$Type": "DomainModels$EntityImpl",
      "Attributes": attr_20,
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
