---

layout: post
title:  "Avoid using system storage objects directly"
categories: Datamodel
prio: 
rulenumber: 42
rulename: 
ruleset: 

---

**Why**
Always inherit for filedocuments and images. Never implement direct assocations to the System Domain Model, because of limits on the configuration of security.

![42.png]({{ site.url }}/assets/42.png)

**How to fix**
Remove direct associations with the System Domain Model. Use inheritance instead (i.e. Generalization in the entity properties).