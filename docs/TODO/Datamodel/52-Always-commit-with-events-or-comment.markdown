---

layout: post
title:  "Always commit with events or comment"
categories: Datamodel
prio: 
rulenumber: 52
rulename: 
ruleset: 

---

**Why**
In principle, always commit with events. When it's necessary to commit without events, always add a comment explaining why events are avoided, and how integrity is guaranteed. A valid reason for not executing events can be performance. But an alternative must be implemented in the applicable flow. 

![52.png]({{ site.url }}/assets/52.png)

**How to fix**
In the Commit activity, select YES for "With events". In case of a valid reason for NO, add an annotation to the flow, elaborating the alternative for data integrity. 