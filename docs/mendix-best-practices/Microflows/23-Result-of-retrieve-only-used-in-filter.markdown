---

layout: post
title:  "Result of retrieve only used in filter"
categories: Microflows
prio: 2
rulenumber: 23
rulename: RetrieveResultOnlyUsedInFilter
ruleset: Performance

---

**Why**
A filter is used on a retrieve action but the list is not used again. This can be a performance killer because Mendix is retrieving all your objects in memory because the filter is modeled a a separate action.

![23.png]({{ site.url }}/assets/23.png)

**How to fix**
Add the filter to the xPath constraint property instead of filtering. Now you will directly get the list of objects that's needed.