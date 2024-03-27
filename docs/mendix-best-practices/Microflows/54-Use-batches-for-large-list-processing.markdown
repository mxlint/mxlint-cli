---

layout: post
title:  "Use batches for large list processing"
categories: Microflows
prio: 
rulenumber: 54
rulename: 
ruleset: 

---

**Why**
Consider using batches when many objects are committed in one microflow. This can be done by adding objects to a list, and committing the list at the end of the microflow. Mendix will then optimize this. Depending on the executed functionality, it is best practice to work with lists of 2.000 to 10.000 objects maximum. The heavier the action being executed, the lower the maximum. Below is an example on how this can be implemented. 

![54.png]({{ site.url }}/assets/54.png)

**How to fix**
