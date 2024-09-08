---

layout: post
title:  "Dataview with only child = dataview"
categories: Forms
prio: 4
rulenumber: 19
rulename: DataviewWithOnlyADataview
ruleset: Performance

---

**Why**
Multiple nested dataviews instead of using an N-deep path creates unnecessary calls your form to the Mendix Runtime. This will affect page loading performance negatively.

![19.png]({{ site.url }}/assets/19.png)

**How to fix**
Put ... in the child dataview directly and retrieve over multiple associations.