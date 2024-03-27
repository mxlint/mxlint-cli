---

layout: post
title:  "Inheritance except from system module"
categories: Datamodel
prio: 4
rulenumber: 8
rulename: AvoidInheritanceExceptSystemModule
ruleset: Performance

---

**Why**
Inheritance, except from system module, is strongly discouraged because of the negative performance side effects.

![8.png]({{ site.url }}/assets/8.png)

**How to fix**
Instead of inheritance, just use separate objects which are associated to the main object. As an alternative, you can add the child’s attributes to the super entity and add an ObjectType enumeration.