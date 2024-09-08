---

layout: post
title:  "Reference selector with editable Never"
categories: Forms
prio: 2
rulenumber: 21
rulename: ReferenceSelectorWithEditableNever
ruleset: Performance

---

**Why**
In certain versions of Mendix a read-only reference selector will try to load the object list as well. This affects performance in a negative way.

![21.png]({{ site.url }}/assets/21.png)

**How to fix**
Use a text box instead, set to read-only.