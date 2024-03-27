---

layout: post
title:  "Make use of SaveObjectX microflows"
categories: Datamodel
prio: 
rulenumber: 50
rulename: 
ruleset: 

---

**Why**
For every entity, there should in a microflow be only one action activity that commits the object. This includes any sub-microflows. Often, this action activity is at the end of the flow. 

![50.png]({{ site.url }}/assets/50.png)

**How to fix**
Remove Commit Object(s) activities if there are more than one in any flow, until there is only one. Preferably, the remaining one is at the end of the flow. 