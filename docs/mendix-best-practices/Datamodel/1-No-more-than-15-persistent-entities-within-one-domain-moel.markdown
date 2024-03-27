---

layout: post
title:  "No more than 15 persistent entities within one domain model"
categories: Datamodel
prio: 3
rulenumber: 1
rulename: NumberOfEntities
ruleset: Maintainability

---

**Why**
The bigger the domain models, the harder they will be to maintain. It adds complexity to your security model as well. The smaller the modules, the easier to reuse.

![1.png]({{ site.url }}/assets/1.png)

**How to fix**
Split domain model into multiple modules.