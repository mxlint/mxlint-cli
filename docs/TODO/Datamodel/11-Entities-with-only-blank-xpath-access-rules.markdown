---

layout: post
title:  "Entities with only blank xpath access rules"
categories: Datamodel
prio: 2
rulenumber: 11
rulename: 
ruleset: Security

---

**Why**
And entity should not only have blank access rules. Most of the times, this means the configuration is not correct, not secure. Even if you put the XPath on your forms, this will not be secure.

![11.png]({{ site.url }}/assets/11.png)

**How to fix**
Make sure every entity has contrained access rules, and if not, correct them.