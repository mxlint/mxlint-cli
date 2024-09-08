---

layout: post
title:  "Retrieve by association combined with check on object != empty"
categories: Microflows
prio: 4
rulenumber: 26
rulename: RetrieveByAssocEmptyCheck
ruleset: Performance

---

**Why**
Retrieving the assocation will most of the times result in a database query. Just checking that a reference != empty can already be done without retrieving the associated object. This is because Mendix is quering the references by default.

![26.png]({{ site.url }}/assets/26.png)

**How to fix**
Remove the retrieve by association and change your split to $mainobject/associationname != empty.