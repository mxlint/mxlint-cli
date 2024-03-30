---

layout: post
title:  "Before and after commit actions that commits the object it self with events"
categories: Datamodel
prio: 1
rulenumber: 10
rulename: AvoidCommitInBeforeAndAfterCommitAction
ruleset: Error

---

**Why**
Before and after commit actions should not commit the object itself with an event. The result is an infinite loop in the microflow.

![10.png]({{ site.url }}/assets/10.png)

**How to fix**
Don't commit in BCo actions and commit without events in ACo actions.