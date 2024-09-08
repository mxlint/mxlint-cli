---

layout: post
title:  "Result of a retrieve multiple times used in aggregates but not used elsewhere"
categories: Microflows
prio: 2
rulenumber: 24
rulename: UseOneAggregatePerRetrieve
ruleset: Performance

---

**Why**
Using one list retrieved from database within multiple aggregates has a negative performance impact. This is because Mendix Runtime will optimize a retrieve action + aggregate action to one aggregate query (e.g. SELECT SUM(..) FROM ..)) instead, retrieving the complete list in memory and a (Java) memory aggregation.

![24.png]({{ site.url }}/assets/24.png)

**How to fix**
For each aggregate, add a new database retrieve action.  This will result in more activities but will be much faster.