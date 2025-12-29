# 🧠 Active Expiration in Redis

Redis uses **Active Expiration** to proactively remove keys that have exceeded their TTL (Time-To-Live), even if they haven't been accessed. This helps free memory and maintain performance.

---

## 🔁 How It Works

- Redis runs a **background job (cron-like)** every **100ms**.
- In each cycle:
  - **Samples 20 keys** with expiration time from each database.
  - **Checks if they are expired**.
  - **Deletes expired keys**.
  - If the percentage of expired keys > 10%, it **repeats the cycle**.

Every 100ms:
- Sample 20 keys with TTL  
- Check expiration  
- Delete expired keys  
- If expired > 10% → repeat  

---

## ⚙️ Configuration: `active-expire-effort`

You can tune Redis's expiration aggressiveness using the `active-expire-effort` parameter (range: 1–10). Higher values mean more aggressive cleanup but may increase CPU usage and latency.

### 🔧 Base Values

| Parameter                          | Base Value | Description |
|-----------------------------------|------------|-------------|
| `ACTIVE_EXPIRE_CYCLE_KEYS_PER_LOOP` | 20         | Keys sampled per DB loop |
| `ACTIVE_EXPIRE_CYCLE_FAST_DURATION` | 1000 µs    | Max duration of fast cycle |
| `ACTIVE_EXPIRE_CYCLE_SLOW_TIME_PERC` | 25%        | Max CPU usage for slow cycle |
| `ACTIVE_EXPIRE_CYCLE_ACCEPTABLE_STALE` | 10%     | Max % of expired keys tolerated |

### 🧮 Formula to Adjust Values

Keys per loop = base + (base / 4 * (effort - 1))  
Fast duration = base + (base / 4 * (effort - 1))  
Slow time %   = base + (2 * (effort - 1))  
Stale %       = base - (effort - 1)  

---

## ✅ Trade-offs

| Effort Level | Expiration Aggressiveness | CPU Usage | Latency Risk |
|--------------|----------------------------|-----------|--------------|
| Low (1)      | Minimal                    | Low       | Minimal      |
| High (10)    | Aggressive                 | High      | Possible     |

---

## 🔗 References

- [Redis EXPIRE Command](https://valkey.io/commands/expire/)
- [Redis Source Code on GitHub](https://github.com/redis/redis/blob/unstable/src/expire.c)
