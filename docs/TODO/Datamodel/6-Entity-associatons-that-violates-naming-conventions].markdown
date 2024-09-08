---

layout: post
title:  "Entity associatons that violates naming conventions]"
categories: Datamodel
prio: 5
rulenumber: 6
rulename: InconsistentAssociationName
ruleset: Convention

---

**Why**
An association name must have the following format: <owner>_<child>[_<description>]

![6.png]({{ site.url }}/assets/6.png)

**How to fix**
Rename the association so it always includes the parent and child domain names. This will make it a lot easier to autocomplete association names within microflows and to use it in modules such as Reporting for Mendix and Model Reflection.