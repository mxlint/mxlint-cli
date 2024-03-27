---

layout: post
title:  "Inheritance from Administration.Account"
categories: Datamodel
prio: 4
rulenumber: 5
rulename: AvoidInheritanceAdministrationAccount
ruleset: Performance

---

**Why**
There is no need to inherit from administration.account. Administration.account may simply be extended, this is not a system module. Avoid unnecessary inheritance as this has a negative effect on performance.

![5.png]({{ site.url }}/assets/5.png)

**How to fix**
Just inherit from system.user instead or adapt Administration.Account so it fits your needs.