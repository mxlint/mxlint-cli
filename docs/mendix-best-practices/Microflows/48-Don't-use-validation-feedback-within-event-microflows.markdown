---

layout: post
title:  "Don't use validation feedback within event microflows"
categories: Microflows
prio:
rulenumber: 48
rulename:
ruleset:

---

**Why**
Don't use feedback activities in Event microflows (after create/after commit etc.)

**How to fix**
Use separate submicroflows and execute them in your microflows triggered by a user action
