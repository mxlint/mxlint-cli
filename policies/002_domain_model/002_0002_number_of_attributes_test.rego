package app.mendix.domain_model.number_of_attributes
import rego.v1


# Test data
attribute1 = {
    "Name": "Attribute1"
}

entity_attr_0 = {
    "Name": "Entity1",
    "Attributes": null,
}

entity_attr_1 = {
    "Name": "Entity1",
    "Attributes": [
        attribute1
    ]
}

forty := numbers.range(1, 40)
attributes_40 = [ 
    { "Name": attribute1.Name }  | n := forty[_]
]

entity_1_attr_40 = {
    "Name": "Entity1",
    "Attributes": attributes_40,
}


# Test cases
test_no_entities if {
	allow with input as {"Entities": null}
}

test_1_entity_1_attribute if {
	allow with input as {"Entities": [entity_attr_1]}
}

test_2_entities if {
	allow with input as {"Entities": [entity_attr_1, entity_attr_1]}
}

test_3_entities_1_empty if {
	allow with input as {"Entities": [entity_attr_1, entity_attr_1, entity_attr_0]}
}

test_1_entity_40_attributes_not_allowed if {
	not allow with input as {"Entities": [entity_1_attr_40]}
}

test_2_entity_40_attributes_1_empty_not_allowed if {
	not allow with input as {"Entities": [entity_1_attr_40, entity_1_attr_40, entity_attr_0]}
}