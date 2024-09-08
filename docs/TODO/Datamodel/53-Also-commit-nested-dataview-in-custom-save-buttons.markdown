---

layout: post
title:  "Also commit nested dataview in custom save buttons"
categories: Datamodel
prio:
rulenumber: 53
rulename:
ruleset:

---

**Why**
 custom save buttons when using nested dataviews via association (!) must be retrieved in save microflows so the objects can be saved.

![53.png]({{ site.url }}/assets/53.png)

**How to fix**
Retrieve and save the objects by retrieve the associated objects.
