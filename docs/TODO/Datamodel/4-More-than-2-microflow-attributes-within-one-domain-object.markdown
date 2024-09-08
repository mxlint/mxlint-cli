---

layout: post
title:  "More than 2 microflow attributes within one domain object"
categories: Datamodel
prio: 4
rulenumber: 4
rulename: AvoidTooMuchCalculatedAttributes
ruleset: Performance

---

**Why**
Microflow attributes have a strong negative impact on performance, especially the attributes that have database interaction. If you managed to eliminate for example about 100 MF attributes on a medium/large application (model size +- 100MB) you will probably increase performance on load tests with factor 3-6.

![4.png]({{ site.url }}/assets/4.png)

**How to fix**
Make the attributes 'stored' and calculate them in an event, or put the microflow attributes in a separate object.