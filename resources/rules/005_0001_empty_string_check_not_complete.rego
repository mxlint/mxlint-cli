# METADATA
# scope: package
# title: Empty String check not complete
# description: Technically, there is a difference between empty and "". Make sure to check them both.
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: Error
#  rulename: EmptyStringCheckNotComplete
#  severity: MEDIUM
#  rulenumber: 005_0001
#  remediation: Always check a string for empty based on != empty and != "". The first one equals database NULL value, the latter one indicates a truncated string.
#  input: .*\$Microflow\.yaml
package app.mendix.microflows.empty_string_check_not_complete
import rego.v1
annotation := rego.metadata.chain()[1].annotations

default allow := false
allow if count(errors) == 0

errors contains error if {
    [p, v] := walk(input)
    last := array.slice(p, count(p) - 1, count(p))[0]
    last ==  "Expression"
    contains(replace(v, " ", ""), "!=''")
    not contains(replace(v, " ", ""), "!=empty")
    error := sprintf("[%v, %v, %v] Expression in Microflow '%v' has incomplete empty string check '%v'",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            input.Name,
            v,
        ]
    )
}
