---

layout: post
title:  "Amount of cross module associations > 3 to another module (check per module!)"
categories: Datamodel
prio: 4
rulenumber: 9
rulename: AvoidExcessiveCrossModuleAssociations
ruleset: Reuseability

---

**Why**
Module x is strongly dependent of module y. This will make it harder to reuse this module.

![9.png]({{ site.url }}/assets/9.png)

**How to fix**
Consider to combine the modules or create a façade within module y.