# Codecrafters – Build Your Own Redis  
## Full Testcases (97 Stages)

> Format: RESP protocol (`\\r\\n`), input/output mô phỏng client ↔ server.

---

## MODULE 1: TCP & RESP BASIC (1–10)

> This module validates RESP parsing, TCP behavior, concurrency, and core command semantics.

---

### 1. Bind to Port

#### Test 1.1 – Default Port (Happy)
- Condition: Server starts without explicit port argument
- Expected: TCP server listens on `0.0.0.0:6379`

#### Test 1.2 – Port Already in Use (Error)
- Condition: Another process already binds to 6379
- Expected: Server fails with non-zero exit or error log

---

### 2. PING

#### Test 2.1 – Basic PING (Happy)
Input:
```
*1
$4
PING

```
Output:
```
+PONG

```

#### Test 2.2 – PING with Message (Happy)
Input:
```
*2
$4
PING
$5
hello

```
Output:
```
+hello

```

#### Test 2.3 – Invalid Arity (Error)
Input:
```
*3
$4
PING
$1
a
$1
b

```
Output:
```
-ERR wrong number of arguments for 'ping' command

```

---

### 3. Concurrent Clients

#### Test 3.1 – Parallel PING (Concurrency)
- Condition: 50 concurrent TCP clients send PING simultaneously
- Expected: All clients receive `+PONG` without blocking or disconnect

#### Test 3.2 – Mixed Commands (Concurrency)
- Clients issue PING / SET / GET concurrently
- Expected: No race conditions, all responses correct

---

### 4. ECHO

#### Test 4.1 – Simple ECHO (Happy)
Input:
```
*2
$4
ECHO
$3
abc

```
Output:
```
$3
abc

```

#### Test 4.2 – Empty String (Edge)
Input:
```
*2
$4
ECHO
$0


```
Output:
```
$0


```

#### Test 4.3 – Invalid Arity (Error)
Input:
```
*1
$4
ECHO

```
Output:
```
-ERR wrong number of arguments for 'echo' command

```

---

### 5. SET

#### Test 5.1 – Simple SET (Happy)
Input:
```
*3
$3
SET
$3
key
$5
value

```
Output:
```
+OK

```

#### Test 5.2 – Overwrite Existing Key (Edge)
Input:
```
*3
$3
SET
$3
key
$3
new

```
Output:
```
+OK

```

#### Test 5.3 – Invalid Arity (Error)
Input:
```
*2
$3
SET
$3
key

```
Output:
```
-ERR wrong number of arguments for 'set' command

```

---

### 6. GET

#### Test 6.1 – Existing Key (Happy)
Input:
```
*2
$3
GET
$3
key

```
Output:
```
$3
new

```

#### Test 6.2 – Missing Key (Edge)
Input:
```
*2
$3
GET
$7
missing

```
Output:
```
$-1

```

#### Test 6.3 – WRONGTYPE (Error)
Input:
```
*2
$3
LPUSH
$3
key

```
Output:
```
-WRONGTYPE Operation against a key holding the wrong kind of value

```

---

### 7. RESP Protocol Robustness

#### Test 7.1 – Partial Frame (Edge)
- Client sends half of RESP frame, pauses, then sends rest
- Expected: Server buffers input and replies correctly

#### Test 7.2 – Malformed Bulk Length (Error)
Input:
```
$X
abc

```
Output:
```
-ERR Protocol error: invalid bulk length

```

---

### 8. Unknown Command

#### Test 8.1 – Unknown Command (Error)
Input:
```
*1
$7
FOOBAR

```
Output:
```
-ERR unknown command 'foobar'

```

---

### 9. Multiple Commands Per Connection

#### Test 9.1 – Sequential Commands (Happy)
Input:
```
*1
$4
PING
*1
$4
PING

```
Output:
```
+PONG
+PONG

```

---

### 10. Graceful Client Disconnect

#### Test 10.1 – Client Close Socket
- Client closes TCP connection mid-session
- Expected: Server cleans up resources without crash

---


## MODULE 2: EXPIRY & RDB PERSISTENCE (11–25)

> Module này kiểm tra **expiry chính xác theo ms**, **negative cases**, và **RDB parsing byte-level** đúng với Codecrafters.

---

### 11. SET PX / EX

#### Test 11.1 – SET PX (Happy)
Input (RESP):
```
*5
$3
SET
$1
k
$1
v
$2
PX
$3
100

```
Output:
```
+OK

```

#### Test 11.2 – SET EX (Happy)
Input:
```
SET k v EX 1
```
Output:
```
+OK
```

#### Test 11.3 – PX Non-integer (Error)
Input:
```
SET k v PX foo
```
Output:
```
-ERR PX value is not an integer or out of range

```

#### Test 11.4 – Wrong Arity
Input:
```
SET k v PX
```
Output:
```
-ERR syntax error

```

---

### 12. GET + Expiry Timing

#### Test 12.1 – Expired Key (Timing Accurate)
- t=0ms: `SET k v PX 50`
- t=55ms: `GET k`

Output:
```
$-1

```

#### Test 12.2 – Boundary (Edge)
- t=0ms: `SET k v PX 50`
- t=49ms: `GET k`

Output:
```
$1
v

```

---

### 13. Lazy Expiry

#### Test 13.1 – Expiry on Access
- Key expired but not actively deleted
- Deleted when `GET` called

---

### 14. Overwrite Resets TTL

#### Test 14.1 – TTL Reset
```
SET k v PX 100
SET k v2
```
- Expected: key does **not expire**

---

### 15. CONFIG GET dir

#### Test 15.1 – Valid Key
Input:
```
*3
$6
CONFIG
$3
GET
$3
dir

```
Output:
```
*2
$3
dir
$<len>
<path>

```

#### Test 15.2 – Unknown Config
Input:
```
CONFIG GET foo
```
Output:
```
*0

```

---

### 16. CONFIG GET dbfilename

#### Test 16.1 – Default Value
Output:
```
*2
$10
dbfilename
$8
dump.rdb

```

---

### 17. RDB Header

#### Test 17.1 – Valid Header
Bytes:
```
52 45 44 49 53 30 30 30 39
```
- ASCII: `REDIS0009`

#### Test 17.2 – Invalid Header (Error)
- Any mismatch → abort load

---

### 18. RDB String Encoding

#### Test 18.1 – Short String
- 6-bit length encoding

#### Test 18.2 – Long String
- 14/32-bit length encoding

---

### 19. RDB Key-Value Load

#### Test 19.1 – Simple KV
- After restart: `GET key` returns value

---

### 20. RDB Expiry Opcode (Seconds)

#### Test 20.1 – Opcode 0xFD
- Expiry seconds < now → skip key

---

### 21. RDB Expiry Opcode (Milliseconds)

#### Test 21.1 – Opcode 0xFC
- ms precision preserved

---

### 22. RDB SELECTDB

#### Test 22.1 – Opcode 0xFE
- Switch DB index

---

### 23. RDB EOF

#### Test 23.1 – Opcode 0xFF
- Parsing stops cleanly

---

### 24. Corrupted RDB

#### Test 24.1 – Truncated File
- Server refuses to start

---

### 25. Persistence End-to-End

#### Test 25.1 – Restart Restore
```
SET a b
SHUTDOWN
RESTART
GET a
```
Output:
```
$1
b

```

---


### 11. SET PX

#### Test 11.1 – PX Expiry (Happy)
Input:
```
*5
$3
SET
$1
k
$1
v
$2
PX
$3
100

```
Output:
```
+OK

```

#### Test 11.2 – Invalid PX Value (Error)
Input:
```
SET k v PX -1
```
Output:
```
-ERR PX value is not an integer or out of range

```

---

### 12. GET After Expiry

#### Test 12.1 – Key Expired (Timing)
t=0ms: SET k v PX 50  
t=60ms: GET k

Output:
```
$-1

```

#### Test 12.2 – Key Still Alive (Edge)
t=0ms: SET k v PX 100  
t=50ms: GET k

Output:
```
$1
v

```

---

### 13. CONFIG GET dir

#### Test 13.1 – Get dir (Happy)
Input:
```
*3
$6
CONFIG
$3
GET
$3
dir

```
Output:
```
*2
$3
dir
$<len>
<path>

```

---

### 14. CONFIG GET dbfilename

#### Test 14.1 – Get dbfilename
Input:
```
CONFIG GET dbfilename
```
Output:
```
*2
$10
dbfilename
$8
dump.rdb

```

---

### 15. RDB Header

#### Test 15.1 – Valid Header (Happy)
- RDB starts with `REDIS0009`
- Expected: Header accepted

#### Test 15.2 – Invalid Header (Error)
- Header mismatch
- Expected: RDB load aborted

---

### 16. RDB String Key

#### Test 16.1 – Load Simple KV
- RDB contains string key/value
- Expected: Key available after restart

---

### 17. RDB Expiry (Seconds)

#### Test 17.1 – Unix Seconds Expiry
- Expiry encoded in seconds
- Expected: Key expires correctly

---

### 18. RDB Expiry (Milliseconds)

#### Test 18.1 – Unix Milliseconds Expiry
- Expiry encoded in ms
- Expected: Precise expiry

---

### 19. RDB Opcode EOF

#### Test 19.1 – EOF Handling
- Opcode `0xFF`
- Expected: Stop parsing cleanly

---

### 20. RDB Opcode SELECTDB

#### Test 20.1 – DB Switch
- Opcode `0xFE`
- Expected: Subsequent keys in selected DB

---

### 21. RDB Opcode EXPIRETIME

#### Test 21.1 – Seconds Expiry Opcode
- Opcode `0xFD`

---

### 22. RDB Opcode EXPIRETIME_MS

#### Test 22.1 – Milliseconds Expiry Opcode
- Opcode `0xFC`

---

### 23. Corrupted RDB

#### Test 23.1 – Truncated File (Error)
- Expected: Server fails to load RDB safely

---

### 24. Multiple Keys Load

#### Test 24.1 – Load Many Keys
- Expected: All keys present

---

### 25. Persistence End-to-End

#### Test 25.1 – Restart Restore
- SET key
- Shutdown
- Restart
- GET key → value

---


## MODULE 3: REPLICATION – MASTER / SLAVE (26–50)

> This module validates Redis-style replication, handshake, command propagation, offsets, and fault tolerance.

---

### 26. INFO replication

#### Test 26.1 – Master Role (Happy)
Input:
```
*2
$4
INFO
$11
replication

```
Output (contains):
```
role:master
```

#### Test 26.2 – Slave Role
- Server started with `--replicaof`
- Output contains `role:slave`

---

### 27. Replica Handshake – PING

#### Test 27.1 – Initial PING
- Replica connects to master
Input:
```
PING
```
Output:
```
+PONG
```

---

### 28. Replica Handshake – REPLCONF listening-port

#### Test 28.1 – Send Listening Port
Input:
```
REPLCONF listening-port 6380
```
Output:
```
+OK
```

---

### 29. Replica Handshake – REPLCONF capa

#### Test 29.1 – Capability Exchange
Input:
```
REPLCONF capa eof
```
Output:
```
+OK
```

---

### 30. PSYNC Handshake

#### Test 30.1 – Full Resync (Happy)
Input:
```
PSYNC ? -1
```
Output:
```
+FULLRESYNC <replid> 0
```

---

### 31. RDB Transfer

#### Test 31.1 – Empty RDB
Output:
```
$<hex_len>
<RDB_BYTES>
```
- Replica loads DB successfully

---

### 32. Command Propagation

#### Test 32.1 – SET Propagation
- Master receives:
```
SET k v
```
- Replica receives same command

#### Test 32.2 – Multiple Commands Order
- Commands replicated in original order

---

### 33. REPLCONF ACK

#### Test 33.1 – ACK Offset
Input:
```
REPLCONF ACK <offset>
```
Output:
```
+OK
```

---

### 34. WAIT Command

#### Test 34.1 – Wait for Replicas (Happy)
Input:
```
WAIT 1 1000
```
Output:
```
:1
```

#### Test 34.2 – Timeout (Edge)
- Replica does not ACK
Output:
```
:0
```

---

### 35. Replica Lag

#### Test 35.1 – Delayed ACK
- ACK sent after delay
- WAIT waits correctly

---

### 36. Partial Resync Rejection

#### Test 36.1 – Unsupported PSYNC
Input:
```
PSYNC replid 100
```
Output:
```
+FULLRESYNC
```

---

### 37. Replica Disconnect

#### Test 37.1 – Slave Drop
- Slave disconnects mid-stream
- Master continues serving clients

---

### 38. Replica Reconnect

#### Test 38.1 – Re-handshake
- Slave reconnects
- Full resync performed

---

### 39. Multiple Replicas

#### Test 39.1 – Fan-out Replication
- Master replicates to N replicas

---

### 40. Replica Read-Only

#### Test 40.1 – Write on Replica (Error)
Input:
```
SET a b
```
Output:
```
-READONLY You can't write against a read only replica

```

---

### 41–45. Fault Tolerance

#### Test 41.1 – Network Partition
- Replica temporarily unreachable

#### Test 42.1 – Resume After Partition
- Full resync triggered

---

### 46–50. Replication Edge Cases

- Offset mismatch
- Duplicate ACKs
- Concurrent replica joins
- Master restart

---


## MODULE 4: STREAMS (51–75)

### 51. XADD (Fixed ID)
**Input**
```
XADD mystream 0-1 field value
```
**Output**
```
$3\r\n0-1\r\n
```

### 52. XADD (Auto ID)
**Input**
```
XADD mystream * field value
```
**Output**
```
$<len>\r\n<ms-seq>\r\n
```

### 53. XADD Validation
**Input**
```
XADD mystream 0-1 field value
```
**Output**
```
-ERR The ID specified is equal or smaller than the target stream top item
```

### 54. XRANGE
**Input**
```
XRANGE mystream - +
```
**Output**
```
*<n>
```

### 55. XREAD
**Input**
```
XREAD STREAMS mystream 0-0
```
**Output**
```
*1
```

### 56. XREAD BLOCK
**Input**
```
XREAD BLOCK 1000 STREAMS mystream $
```
**Output**
```
*1
```

### 57–65. MAXLEN
- Trim stream size.

### 66–70. XGROUP CREATE
- Create consumer groups.

### 71–75. XREADGROUP
- Read as consumer.

---

## MODULE 5: TRANSACTIONS (76–85)

### 76. MULTI
**Input**
```
MULTI
```
**Output**
```
+OK
```

### 77. Queue Commands
**Input**
```
SET a 1
```
**Output**
```
+QUEUED
```

### 78. EXEC
**Input**
```
EXEC
```
**Output**
```
*1\r\n+OK\r\n
```

### 79. DISCARD
**Input**
```
DISCARD
```
**Output**
```
+OK
```

### 80–85. Atomicity
- Either all commands succeed or none.

---

## MODULE 6: ADVANCED DATA TYPES (86–97)

### 86. ZADD
**Input**
```
ZADD z 1 one
```
**Output**
```
:1
```

### 87. ZRANGE
**Input**
```
ZRANGE z 0 -1
```
**Output**
```
*1\r\n$3\r\none\r\n
```

### 88. ZSCORE
**Input**
```
ZSCORE z one
```
**Output**
```
$1\r\n1\r\n
```

### 89–90. ZSET Update
- Update existing scores.

### 91. HSET
**Input**
```
HSET h f v
```
**Output**
```
:1
```

### 92. HGET
**Input**
```
HGET h f
```
**Output**
```
$1\r\nv\r\n
```

### 93. HGETALL
**Output**
```
*2
```

### 94. Hash Missing Field
**Output**
```
$-1
```

### 95. SADD
**Input**
```
SADD s a
```
**Output**
```
:1
```

### 96. SISMEMBER
**Input**
```
SISMEMBER s a
```
**Output**
```
:1
```

### 97. Lists (LPUSH / LPOP)
**Input**
```
LPUSH l a
```
**Output**
```
:1
```



## MODULE 4: STREAMS – XADD / XRANGE / XREAD (51–75)

> Module này kiểm tra **Redis Streams chuẩn**, bao gồm ID generation, blocking reads, trimming, consumer groups và lỗi.

---

### 51. XADD – Fixed ID

#### Test 51.1 – Add with Explicit ID (Happy)
Input:
```
*5
$4
XADD
$1
s
$3
0-1
$1
f
$1
v

```
Output:
```
$3
0-1

```

#### Test 51.2 – ID Smaller Than Last (Error)
Input:
```
XADD s 0-0 f v
```
Output:
```
-ERR The ID specified is equal or smaller than the target stream top item
```

---

### 52. XADD – Auto ID (*)

#### Test 52.1 – Auto ID Generation
Input:
```
XADD s * f v
```
Output:
```
$<len>
<ms>-0

```

#### Test 52.2 – Same ms increments seq
- Two XADD within same millisecond
- IDs: `<ms>-0`, `<ms>-1`

---

### 53. XADD – Invalid Arguments

#### Test 53.1 – Missing Field Value
Input:
```
XADD s * f
```
Output:
```
-ERR wrong number of arguments for 'xadd' command
```

---

### 54. XRANGE

#### Test 54.1 – Full Range
Input:
```
XRANGE s - +
```
Output:
```
*1
*2
$<id>
*2
$1 f
$1 v
```

#### Test 54.2 – Empty Stream
Output:
```
*0
```

---

### 55. XRANGE – Boundaries

#### Test 55.1 – Start Exclusive
```
XRANGE s (0-1 +
```

---

### 56. XREAD

#### Test 56.1 – Simple Read
Input:
```
XREAD STREAMS s 0-0
```
Output:
```
*1
*2
$1 s
*1
*2
$<id>
*2
$1 f
$1 v
```

---

### 57. XREAD – BLOCK

#### Test 57.1 – Blocking Read Returns Data
Input:
```
XREAD BLOCK 1000 STREAMS s $
```
- Data arrives within 1000ms

#### Test 57.2 – Timeout
Output:
```
$-1
```

---

### 58. XREAD – Multiple Streams

#### Test 58.1 – Read Two Streams
```
XREAD STREAMS s1 s2 0-0 0-0
```

---

### 59. Stream Trimming – MAXLEN

#### Test 59.1 – Approximate Trim
```
XADD s MAXLEN ~ 2 * f v
```
- Stream length <= 2

---

### 60. Stream Trimming – Exact

#### Test 60.1 – Exact Trim
```
XADD s MAXLEN = 1 * f v
```

---

### 61. XGROUP CREATE

#### Test 61.1 – Create Group
```
XGROUP CREATE s g 0-0
```
Output:
```
+OK
```

#### Test 61.2 – Group Exists (Error)
```
-ERR BUSYGROUP Consumer Group name already exists
```

---

### 62. XREADGROUP

#### Test 62.1 – Read as Consumer
```
XREADGROUP GROUP g c STREAMS s >
```

---

### 63. XREADGROUP – Pending Entries

#### Test 63.1 – Re-read Pending
```
XREADGROUP GROUP g c STREAMS s 0
```

---

### 64. XACK

#### Test 64.1 – Acknowledge Message
```
XACK s g <id>
```
Output:
```
:1
```

---

### 65. XPENDING

#### Test 65.1 – Pending Summary
```
XPENDING s g
```

---

### 66. Blocking XREADGROUP

#### Test 66.1 – Block Until Message
```
XREADGROUP GROUP g c BLOCK 1000 STREAMS s >
```

---

### 67–70. Stream Edge Cases

- Deleting stream
- Reading from non-existent stream
- Wrong ID format
- Wrong argument order

---

### 71–75. Negative Tests

- WRONGTYPE on non-stream key
- ERR syntax error
- ERR invalid stream ID
- BLOCK without timeout

---



## MODULE 5: TRANSACTIONS – MULTI / EXEC / DISCARD (76–85)

> Module này kiểm tra **transaction queue**, **atomicity**, **error handling** và RESP output chính xác.

---

### 76. MULTI

#### Test 76.1 – Start Transaction
Input:
```
MULTI
```
Output:
```
+OK
```

---

### 77. Queue Commands

#### Test 77.1 – Queue SET
```
MULTI
SET a 1
```
Output:
```
+QUEUED
```

#### Test 77.2 – Queue GET
```
GET a
```
Output:
```
+QUEUED
```

---

### 78. EXEC (Happy Path)

#### Test 78.1 – Execute Transaction
```
MULTI
SET a 1
GET a
EXEC
```
Output:
```
*2
+OK
$1
1
```

---

### 79. EXEC Atomicity

#### Test 79.1 – All or Nothing
```
MULTI
SET a 1
BADCOMMAND
EXEC
```
Output:
```
-EXECABORT Transaction discarded because of previous errors.
```

---

### 80. Runtime Error Inside EXEC

#### Test 80.1 – WRONGTYPE during EXEC
```
SET a 1
MULTI
LPUSH a x
EXEC
```
Output:
```
*1
-WRONGTYPE Operation against a key holding the wrong kind of value
```

---

### 81. DISCARD

#### Test 81.1 – Discard Transaction
```
MULTI
SET a 1
DISCARD
```
Output:
```
+OK
```

---

### 82. MULTI Misuse

#### Test 82.1 – Nested MULTI (Error)
```
MULTI
MULTI
```
Output:
```
-ERR MULTI calls can not be nested
```

---

### 83. EXEC Without MULTI

#### Test 83.1 – Error
```
EXEC
```
Output:
```
-ERR EXEC without MULTI
```

---

### 84. DISCARD Without MULTI

#### Test 84.1 – Error
```
DISCARD
```
Output:
```
-ERR DISCARD without MULTI
```

---

### 85. WATCH (Optional Edge)

#### Test 85.1 – WATCH Not Implemented
```
WATCH a
```
Output:
```
-ERR WATCH is not supported
```

---

## MODULE 6: ADVANCED DATA TYPES (86–97)

> Module này kiểm tra **ZSET, HASH, SET, LIST**, đúng kiểu dữ liệu, RESP và lỗi.

---

### 86. ZADD

#### Test 86.1 – Add Member
```
ZADD z 1 one
```
Output:
```
:1
```

---

### 87. ZRANGE

#### Test 87.1 – Range Ascending
```
ZRANGE z 0 -1
```
Output:
```
*1
$3
one
```

---

### 88. ZSCORE

#### Test 88.1 – Get Score
```
ZSCORE z one
```
Output:
```
$1
1
```

---

### 89. ZADD Update

#### Test 89.1 – Update Score
```
ZADD z 2 one
```
Output:
```
:0
```

---

### 90. ZSET WRONGTYPE

#### Test 90.1 – Wrong Type Error
```
SET a 1
ZADD a 1 x
```
Output:
```
-WRONGTYPE Operation against a key holding the wrong kind of value
```

---

### 91. HSET

#### Test 91.1 – Add Field
```
HSET h f v
```
Output:
```
:1
```

---

### 92. HGET

#### Test 92.1 – Get Field
```
HGET h f
```
Output:
```
$1
v
```

---

### 93. HGETALL

#### Test 93.1 – Get All Fields
```
HGETALL h
```
Output:
```
*2
$1
f
$1
v
```

---

### 94. HASH WRONGTYPE

#### Test 94.1 – Error
```
SET a 1
HSET a f v
```
Output:
```
-WRONGTYPE Operation against a key holding the wrong kind of value
```

---

### 95. SADD

#### Test 95.1 – Add Member
```
SADD s a
```
Output:
```
:1
```

---

### 96. SISMEMBER

#### Test 96.1 – Check Member
```
SISMEMBER s a
```
Output:
```
:1
```

---

### 97. LIST – LPUSH / LPOP

#### Test 97.1 – Push and Pop
```
LPUSH l a
LPOP l
```
Output:
```
$1
a
```

---

