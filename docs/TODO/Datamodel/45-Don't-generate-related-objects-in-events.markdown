---

layout: post
title:  "Don't generate related objects in events"
categories: Datamodel
prio:
rulenumber: 45
rulename:
ruleset:

---

**Why**
Unwanted behaviour if the after create event of a parent object creates child objects when this is in use in nested dataviews. 

**How to fix**
Use a microflow with a Get or Create pattern
