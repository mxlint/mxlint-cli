---

layout: post
title:  "Don't save for UI on change events"
categories: Forms
prio: 
rulenumber: 51
rulename: 
ruleset: 

---

**Why**
When a microflow (e.g. OnChange) is called from a dataview, don't use a Commit activity. The user needs to be able to cancel by clicking the Cancel button.

![51.png]({{ site.url }}/assets/51.png)

**How to fix**
Remove any Commit activities from microflows that are called from dataviews.