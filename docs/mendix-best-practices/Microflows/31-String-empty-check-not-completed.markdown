---

layout: post
title:  "String empty check not completed"
categories: Microflows
prio: 3
rulenumber: 31
rulename: StringEmptyCheckNotComplete
ruleset: Error

---

**Why**
Technically, there is a difference between empty and "". Make sure to check them both.

![31.png]({{ site.url }}/assets/31.png)

**How to fix**
Always check a string for empty based on != empty and != "". The first one equals database NULL value, the latter one indicates a truncated string.