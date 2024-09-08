---

layout: post
title:  "Localize set to false when using DateTime as Date + time"
categories: Datamodel
prio: 
rulenumber: 41
rulename: 
ruleset: 

---

**Why**
When the time of a DateTime attribute *is* relevant, set Localize to YES. This ensures that the rendering on pages and in Retrieve activities goes well, e.g. when you compare the value to [%CurrentDatetime%].

![41.png]({{ site.url }}/assets/41.png)

**How to fix**
Change localization of the DateTime attribute to YES. If you have existing data in your database, don't do this without intelligent conversion.