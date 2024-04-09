# METADATA
# scope: package
# title: More than 20 modules in project
# description: The bigger the application, the harder to maintain.
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: Maintainability
#  rulename: NumberOfModules
#  severity: MEDIUM
#  rulenumber: 003_0001
#  remediation: Consider a multi-app stategy to avoid creating one big (unmaintainable) monstrous application.
#  input: "Metadata.yaml"
package app.mendix.modules.number_of_modules
import rego.v1
annotation := rego.metadata.chain()[1].annotations

default allow := false
allow if count(errors) == 0

max_count := 20

errors contains error if {
    not input.Modules == null
    some i
    user_modules := [item | input.Modules[i].Attributes.FromAppStore == false ; item := input.Modules[i]]
    count_items = count(user_modules)
    count_items > max_count
    error := sprintf("[%v, %v, %v] There are %v user modules which is more than %v",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            count_items,
            max_count
        ]
    )
}
