# METADATA
# title: No more than 15 persistent entities within one domain model
# description: The bigger the domain models, the harder they will be to maintain. It adds complexity to your security model as well. The smaller the modules, the easier to reuse.
# authors:
# - Xiwen Cheng <x@cinaq.com>
# custom:
#  category: security
#  rulename: NumberOfEntities
#  priority: 2
#  rulenumber: 002-0001
#  remediation: Split domain model into multiple modules.
#  input: "*/DomainModels$DomainModel.yaml"

package app.mendix.projectsettings

import rego.v1

default less_than_15_entities := false

less_than_15_entities if {
    input.Entities == null
}
less_than_15_entities if {
    count(input.Entities) <= 15
}