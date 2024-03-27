---

layout: post
title:  "Style property used"
categories: Forms
prio: 5
rulenumber: 13
rulename: StylePropertyUsed
ruleset: Maintainability

---

**Why**
Avoid using the style property, because this will make the life of your UI designer a lot more complicated. It will be harder to overrule styles from CSS file level.

**How to fix**
Use generic classes instead, defined by the theme.