---

layout: post
title:  "Retrieve first instead of list for empty checks"
categories: Microflows
prio:
rulenumber: 46
rulename:
ruleset:

---

**Why**
The use of retrieve action to check if an object exists should always retrieve only the first record instead of the whole list.

![46.png]({{ site.url }}/assets/46.png)

**How to fix**
If the retrieve of the objects is followed by a test if the object != empty then set the retrieve to "first".
