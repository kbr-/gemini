# Failure Analysis

This document describes how to analyze failures reported by Gemini to determine if they are bugs in system under test.

## Failure Types

* A system under test crash.
* A persistent data mismatch between system under test and test oracle.
* A transient data mismatch between system under test and test oracle.

### System crash

System crashes are easiest to analyse because they are always a bug in system under test.

To turn this failure type into a bug report, remember to report:

* System core dump
* Schema used by Gemini
* Gemini trace log (if available)

### Persistent data mismatch

### Transient data mismatch
