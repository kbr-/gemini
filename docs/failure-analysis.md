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

### Data mismatch

The fundamental assumption of Gemini is that both system under test and test oracle are in the same client-visible state after each mutation. When Gemini detects a data _mismatch_, it reports it as follows:

```json
     "result": {
         "write_ops": 11545,
         "write_errors": 0,
         "read_ops": 122,
         "read_errors": 1,
         "errors": [
             {
                 "timestamp": "2019-10-28T08:44:27.494611124Z",
                 "message": "Validation failed: rows differ (-map[ck0:149.100.125.13 ck1:2008-08-10 06:55:16 +0000 UTC ck2:YN7Gq2zkAvdG8FJP784z1hwI8qrKSlE7q74dU16GMTEo9kXE0yEIkX7aExEEtT3HidiV53IEeDWz13gqHoDdfdDHtszF0iUSLF col0:map[128658798019239.737:163.210.75.165 251584688897365.312:123.90.141.66 314040171546363.054:52.201.172.208 652785548908260.195:254.56.50.165 1074685168819392.328:150.50.82.60 6146281011689446.585:244.155.41.92 7049168503758589.494:240.175.53.169] col1[0]:0.5162352993615078 col1[1]:0.3533933 col1[2]:1983-04-19 20:49:20 +0000 UTC col1[3]:1995-03-22 00:00:00 +0000 UTC col1[4]:{0 0 2340000000000} col1[5]:0.4280471268508381 col1[6]:1994-04-05 00:00:00 +0000 UTC col1[7]:[53 54 53 50 53 53 51 53 53 54 51 56 52 102 51 55 52 52 54 101 54 49 53 48 55 50 53 52 51 52 51 53 55 49 51 54 51 57 53 56 52 51 55 54 52 55 51 50 54 100 55 54 54 49 55 97 51 54 51 55 52 57 51 52 54 56 51 48 54 53 54 53 53 48 55 57 54 99 52 98 51 54 50 98 53 54 55 49 53 53 54 55 52 50 52 50 54 98 53 48 51 53 54 49 55 97 54 100 52 99 51 57 55 51 52 49 53 54 51 52 55 49 53 53 54 54 53 54 54 56 52 98 54 49 55 53 52 100 54 55 52 52 55 48 55 56 51 48 52 102 53 55 54 52 51 52 52 56 55 55 54 56 52 98 51 56 53 57 52 51 53 53 54 52 54 49 52 52 54 50 52 49 51 55 53 55 52 98 55 49 51 53 53 52 53 52 53 55 55 51 53 49 55 57 55 52 54 53 53 53 54 52 54 101 53 48 53 49 55 56 51 55 53 49 53 52 54 51 54 102 55 52 55 48 54 49 54 52 52 98 52 55 52 102 50 102 53 48 55 53 54 55 52 50 54 98 52 102 51 56 55 52 55 48 54 100 52 55 55 55 54 50 52 102 52 56 52 100 53 56 55 50 55 54 52 100 54 53 52 97 55 57 53 48 53 55 52 49 52 100 52 53 54 49 54 49 52 54 55 54 54 5....
52 102 55 56 55 56 52 51 51 57 55 56 51 54 53 53 52 55 54 97 54 50 54 53 54 56 52 99 51 48 54 99 51 53 51 55 51 54 53 52 51 54 55 48 53 54 54 52 51 54 54 52 55 51 51 52 53 53 51 54 53 55] col1[8]:Pr019toIsJ2rTKqiAQvLw2QX9k3dih66De2BVxuHck8+tr+6O0 col2:[6dfcf500-5fc1-11e6-adea-12bf508eaf94 f20f5c00-48c3-11c1-adeb-12bf508eaf94 6dfcf500-5fc1-11e6-adea-12bf508eaf94 f20f5c00-48c3-11c1-adeb-12bf508eaf94] col3:[d12dfc00-ebbf-11b6-adec-12bf508eaf94 8b04d280-0df2-11df-aded-12bf508eaf94 77316080-109a-11b6-adee-12bf508eaf94 40b98800-cd5a-11c7-adef-12bf508eaf94 28c5e880-f3af-11e0-adf0-12bf508eaf94 fc0d8480-e9ef-11e8-adf1-12bf508eaf94] col4:1899936722 col5:map[-124:3796881086121027595 -113:4626969566850651375 -77:7718400492645432069 -72:6401329042823927557 -63:5829149512400651502 -22:6162554463243434646 16:183576932513034109 44:6812843939577363700 112:1921272292244960019] pk0:17070878 pk1:8600041440762243304 pk2:-23875 pk3:1473294943 pk4:1098769373]): root[\"col2\"][?->2]:\n\t-: <non-existent>\n\t+: s\"6dfcf500-5fc1-11e6-adea-12bf508eaf94\"\nroot[\"col2\"][?->3]:\n\t-: <non-existent>\n\t+: s\"f20f5c00-48c3-11c1-adeb-12bf508eaf94\"\n",
                 "query": "SELECT * FROM ks1.table1 WHERE pk0=17070878 AND pk1=8600041440762243304 AND pk2=-23875 AND pk3=1473294943 AND pk4=1098769373 "
             }
         ]
```

The first step after a data mismatch failure is to determine if the mismatch is _permanent_ or _transient_.
