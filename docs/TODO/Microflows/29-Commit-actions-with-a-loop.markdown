---

layout: post
title:  "Commit actions with a loop"
categories: Microflows
prio: 3
rulenumber: 29
rulename: AvoidCommitInLoop
ruleset: Performance

---

**Why**
Commiting objects within a loop will fire a SQL Update query for each iteration. This will decrease performance.

![29.png]({{ site.url }}/assets/29.png)

**How to fix**
Consider committing objects outside the loop. Within the loop, add them to a list.