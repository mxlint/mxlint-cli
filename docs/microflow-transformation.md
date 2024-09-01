## Microflow Transformation

This document outlines the thought process and approach on how to transform a program (or function if you want) from a graph like format to as-linear-as-possible textual representation.

### Mendix Microflows are not DAG's

Mendix Microflows are in abstract form a graph structure. For those warry of graph-theory, at first sight most graphs are Directed Acyclic Graph (DAG). Because there is a starting point and multiple end states. There is a minor detail: it's possible to create loops using `Exclusive Split` and `Exclusive Merge` actions. These are comparable to defining `labels` and `goto` in more classical programming languages.


### Goto