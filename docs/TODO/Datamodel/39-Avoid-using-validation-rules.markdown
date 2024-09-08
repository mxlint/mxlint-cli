---

layout: post
title:  "Avoid using validation rules"
categories: Datamodel
prio: 
rulenumber: 39
rulename: 
ruleset: 

---

**Why**
Validation rules on domain model level will give the users unexpected errors. For example, when importing data you maybe want to store invalid data temporary.

![39.png]({{ site.url }}/assets/39.png)

**How to fix**
Remove datamodel validations and validate by microflows from UI. If you really need a validation rule, make sure to document it.