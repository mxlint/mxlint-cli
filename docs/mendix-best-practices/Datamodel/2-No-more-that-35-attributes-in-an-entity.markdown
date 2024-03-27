---

layout: post
title:  "No more that 35 attributes in an entity"
categories: Datamodel
prio: 3
rulenumber: 2
rulename: NumberOfAttributes
ruleset: Maintainability

---

**Why**
The bigger the entities, the slower your application will become when handling the data. This is because Mendix is using SELECT * queries a lot and will retrieve a lot of unnecessary data.

![2.png]({{ site.url }}/assets/2.png)

**How to fix**
Normalize your datamodel. Split your object into multiple objects. If the attributes really belong to each other in a one-to-one relation, just draw a one-to-one relation between the objects.