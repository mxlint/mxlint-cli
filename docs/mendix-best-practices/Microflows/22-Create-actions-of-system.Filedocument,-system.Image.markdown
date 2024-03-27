---

layout: post
title:  "Create actions of system.Filedocument, system.Image"
categories: Microflows
prio: 3
rulenumber: 22
rulename: AvoidCreateFileDocOrImage
ruleset: Security

---

**Why**
Avoid creating filedocuments or images. Because security options on both objects are very limited.

![22.png]({{ site.url }}/assets/22.png)

**How to fix**
Inherit always from system domain objects before using them. In that way you can configure your own security.