---

layout: post
title:  "Delete behaviour in place"
categories: Datamodel
prio: 
rulenumber: 38
rulename: 
ruleset: 

---

**Why**
To avoid data-corruption and floating data make sure there is delete behaviour on every association.

![38.png]({{ site.url }}/assets/38.png)

**How to fix**
Model the delete behaviour. If there is no delete behaviour on purpose, please document it on the owner entity.