# METADATA
# scope: package
# title: Inline style property used
# description: Avoid using the style property, because this will make the life of your UI designer a lot more complicated. It will be harder to overrule styles from CSS file level.
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: Maintainability
#  rulename: InlineStylePropertyUsed
#  severity: MEDIUM
#  rulenumber: 004_0001
#  remediation: Use generic classes instead, defined by the theme.
#  input: "*/**/*$Page.yaml"
package app.mendix.pages.inline_style_property_used
import rego.v1
annotation := rego.metadata.chain()[1].annotations

default allow := false
allow if count(errors) == 0

errors contains error if {
    [p, v] := walk(input)
    # Check if the path ends with "Style" and value is not an empty string
    last := array.slice(p, count(p) - 1, count(p))[0]
    last ==  "Style"
    v != ""
    error := sprintf("[%v, %v, %v] Form with name '%v' has inlined style property with value '%v'",
        [
            annotation.custom.severity,
            annotation.custom.category,
            annotation.custom.rulenumber,
            input.Name,
            v,
        ]
    )
}
