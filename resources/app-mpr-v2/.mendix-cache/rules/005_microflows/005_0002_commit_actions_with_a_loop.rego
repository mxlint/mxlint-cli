# METADATA
# scope: package
# title: Commit actions with a loop
# description: Commiting objects within a loop will fire a SQL Update query for each iteration.
# authors:
# - Viktor Berlov <viktor@cinaq.com>
# custom:
#  category: Microflows
#  rulename: AvoidCommitInLoop
#  severity: MEDIUM
#  rulenumber: 005_0002
#  remediation: Consider committing objects outside the loop. Within the loop, add them to a list.
#  input: "**/*$Microflow.yaml"
package app.mendix.microflows.commit_actions_with_a_loop

import rego.v1

annotation := rego.metadata.chain()[1].annotations

default allow := false

allow if count(errors) == 0

errors contains error if {
	name := input.Name
	main_function := input.MainFunction

	some attr in main_function
	attr.Attributes["$Type"] == "Microflows$LoopedActivity"
	some commit_action in attr.Attributes.ObjectCollection.Objects
	commit_action.Action["$Type"] == "Microflows$CommitAction"

	error := sprintf(
		"[%v, %v, %v] Commit actions inside %v loop",
		[
			annotation.custom.severity,
			annotation.custom.category,
			annotation.custom.rulenumber,
			name,
		],
	)
}

errors contains error if {
	name := input.Name
	main_function := input.MainFunction
	some attr in main_function
	attr.Attributes["$Type"] == "Microflows$LoopedActivity"
	some change_action in attr.Attributes.ObjectCollection.Objects
	change_action.Action["$Type"] == "Microflows$ChangeAction"
	change_action.Action.Commit == "Yes"

	error := sprintf(
		"[%v, %v, %v] Commit set to Yes for Change actions inside %v loop",
		[
			annotation.custom.severity,
			annotation.custom.category,
			annotation.custom.rulenumber,
			name,
		],
	)
}
