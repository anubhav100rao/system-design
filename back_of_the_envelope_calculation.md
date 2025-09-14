# Back-of-the-Envelope Calculations for System Design Interviews

## Table of Contents

1. [Why Back-of-the-Envelope Calculations Matter](#why-calculations-matter)
2. [Key Numbers to Remember](#key-numbers)
3. [Common Calculation Categories](#calculation-categories)
4. [Step-by-Step Approach](#step-by-step-approach)
5. [Real-World Examples](#examples)
6. [Common Pitfalls and Tips](#pitfalls-and-tips)
7. [Practice Problems](#practice-problems)

## Why Back-of-the-Envelope Calculations Matter {#why-calculations-matter}

Back-of-the-envelope calculations serve several critical purposes in system design interviews:

-   **Validate feasibility**: Ensure your proposed solution can handle the expected load
-   **Guide design decisions**: Help choose between different architectural approaches
-   **Demonstrate analytical thinking**: Show interviewers your problem-solving methodology
-   **Estimate resource requirements**: Calculate storage, bandwidth, and compute needs
-   **Identify bottlenecks**: Spot potential performance issues before detailed design

## Key Numbers to Remember {#key-numbers}

### Time Units

-   **1 second** = 1,000 milliseconds (ms)
-   **1 millisecond** = 1,000 microseconds (μs)
-   **1 microsecond** = 1,000 nanoseconds (ns)

### Data Units

-   **1 KB** = 1,000 bytes (10³)
-   **1 MB** = 1,000,000 bytes (10⁶)
-   **1 GB** = 1,000,000,000 bytes (10⁹)
-   **1 TB** = 1,000,000,000,000 bytes (10¹²)
-   **1 PB** = 1,000,000,000,000,000 bytes (10¹⁵)

### Latency Numbers (Jeff Dean's Numbers)

-   **L1 cache reference**: 0.5 ns
-   **Branch mispredict**: 5 ns
-   **L2 cache reference**: 7 ns
-   **Mutex lock/unlock**: 100 ns
-   **Main memory reference**: 100 ns
-   **Compress 1KB with Zippy**: 10,000 ns = 10 μs
-   **Send 1KB over 1 Gbps network**: 10,000 ns = 10 μs
-   **Read 4KB randomly from SSD**: 150,000 ns = 150 μs
-   **Read 1MB sequentially from memory**: 250,000 ns = 250 μs
-   **Round trip within same datacenter**: 500,000 ns = 500 μs
-   **Read 1MB sequentially from SSD**: 1,000,000 ns = 1 ms
-   **HDD seek**: 10,000,000 ns = 10 ms
-   **Read 1MB sequentially from HDD**: 30,000,000 ns = 30 ms
-   **Send packet CA→Netherlands→CA**: 150,000,000 ns = 150 ms

### Throughput Numbers

-   **SSD**: ~100MB/s to 1GB/s
-   **HDD**: ~50-200 MB/s
-   **1 Gbps network**: ~100 MB/s
-   **Memory**: ~10-100 GB/s
-   **CPU**: Billions of operations per second

### Common System Capacities

-   **Single server**: ~10,000-100,000 concurrent connections
-   **Database**: ~1,000-10,000 QPS per core
-   **Web server**: ~10,000-100,000 requests per second
-   **CDN**: ~100,000+ requests per second

## Common Calculation Categories {#calculation-categories}

### 1. Traffic Estimation

Calculate requests per second (QPS/RPS):

```
Daily Active Users × Average requests per user per day ÷ 86,400 seconds
```

**Peak traffic multiplier**: Usually 2-10x average traffic

### 2. Storage Estimation

```
Total Storage = Number of records × Average record size × Replication factor
```

**Growth over time**:

```
Future storage = Current storage × (1 + growth_rate)^years
```

### 3. Bandwidth Estimation

```
Bandwidth = Average request size × Requests per second
```

**Considerations**:

-   Incoming bandwidth (writes)
-   Outgoing bandwidth (reads)
-   Read-to-write ratio (typically 100:1 to 1000:1 for read-heavy systems)

### 4. Memory Estimation

**Cache sizing (80-20 rule)**:

```
Cache size = 20% of daily requests × Average response size
```

**Active user sessions**:

```
Memory for sessions = Concurrent users × Session data size
```

### 5. Server Estimation

```
Number of servers = Total QPS ÷ QPS per server
```

Factor in:

-   CPU utilization target (typically 70-80%)
-   Memory requirements
-   Network I/O limits
-   Redundancy and failover

## Step-by-Step Approach {#step-by-step-approach}

### Step 1: Clarify Requirements

-   **Users**: How many daily/monthly active users?
-   **Scale**: What's the expected growth rate?
-   **Features**: What are the core functionalities?
-   **Performance**: What are the latency requirements?

### Step 2: Estimate Traffic

1. Calculate daily active users (DAU)
2. Estimate requests per user per day
3. Calculate average QPS: `DAU × requests_per_user ÷ 86,400`
4. Calculate peak QPS: `Average QPS × peak_multiplier`

### Step 3: Estimate Storage

1. Identify data types and their sizes
2. Calculate storage per user/request
3. Factor in metadata overhead (usually 20-30%)
4. Consider replication (typically 3x)
5. Plan for future growth

### Step 4: Estimate Bandwidth

1. Calculate read bandwidth: `Read QPS × average_response_size`
2. Calculate write bandwidth: `Write QPS × average_request_size`
3. Add safety margin (typically 20-50%)

### Step 5: Estimate Servers

1. Determine bottleneck (CPU, memory, I/O, network)
2. Calculate servers needed for each bottleneck
3. Use the highest number
4. Add redundancy

## Real-World Examples {#examples}

### Example 1: URL Shortener (like bit.ly)

**Requirements**:

-   100M URLs shortened per month
-   Read-to-write ratio: 100:1
-   URL lifespan: 5 years
-   Average URL length: 100 characters

**Calculations**:

**Traffic**:

-   Write QPS: 100M ÷ (30 × 24 × 3600) = ~38 QPS
-   Read QPS: 38 × 100 = 3,800 QPS
-   Peak QPS: 3,800 × 3 = ~11,400 QPS

**Storage**:

-   URLs per month: 100M
-   Storage per URL: 100 bytes (original) + 7 bytes (short) + 50 bytes (metadata) = ~160 bytes
-   Monthly storage: 100M × 160 bytes = 16 GB
-   5-year storage: 16 GB × 12 × 5 = 960 GB ≈ 1 TB
-   With replication: 1 TB × 3 = 3 TB

**Bandwidth**:

-   Write: 38 QPS × 160 bytes = ~6 KB/s
-   Read: 3,800 QPS × 160 bytes = ~600 KB/s
-   Total: ~606 KB/s

**Servers**:

-   Assuming 1,000 QPS per server: 11,400 ÷ 1,000 = ~12 servers
-   With redundancy: ~15-20 servers

### Example 2: Chat Application (like WhatsApp)

**Requirements**:

-   1B users, 500M daily active
-   Average 40 messages per user per day
-   Average message size: 100 bytes
-   Message retention: 1 year

**Calculations**:

**Traffic**:

-   Daily messages: 500M × 40 = 20B messages
-   Average QPS: 20B ÷ 86,400 = ~230K QPS
-   Peak QPS: 230K × 3 = ~700K QPS

**Storage**:

-   Daily storage: 20B × 100 bytes = 2 TB
-   Yearly storage: 2 TB × 365 = 730 TB
-   With replication: 730 TB × 3 = ~2.2 PB

**Bandwidth**:

-   Average: 230K QPS × 100 bytes = 23 MB/s
-   Peak: 700K QPS × 100 bytes = 70 MB/s

**Servers**:

-   Message servers (assuming 10K QPS each): 700K ÷ 10K = 70 servers
-   Database servers: Based on storage and write load
-   With redundancy and geographic distribution: 200+ servers

### Example 3: Social Media Feed

**Requirements**:

-   300M users, 200M daily active
-   Average 5 posts per user per day (read)
-   Average 0.1 posts per user per day (write)
-   Average post size: 1KB
-   Timeline generation for active users

**Calculations**:

**Traffic**:

-   Read QPS: 200M × 5 ÷ 86,400 = ~11,600 QPS
-   Write QPS: 200M × 0.1 ÷ 86,400 = ~230 QPS
-   Timeline generation QPS: Depends on refresh rate

**Storage**:

-   Daily posts: 200M × 0.1 = 20M posts
-   Daily storage: 20M × 1KB = 20 GB
-   With metadata and indexes: 20 GB × 2 = 40 GB/day
-   Yearly: 40 GB × 365 = ~15 TB
-   With replication: 15 TB × 3 = 45 TB

**Memory (for caching)**:

-   Active user feeds: 50M users × 20 posts × 1KB = 1 TB
-   Hot posts cache: 10M posts × 1KB = 10 GB
-   Total cache: ~1 TB

## Common Pitfalls and Tips {#pitfalls-and-tips}

### Pitfalls to Avoid

1. **Not rounding numbers**: Use round numbers for easier calculation
2. **Forgetting peak load**: Always consider traffic spikes
3. **Ignoring metadata**: Database overhead is typically 20-30%
4. **Not considering replication**: Most systems use 3x replication
5. **Mixing up units**: Be careful with KB vs MB vs GB
6. **Overlooking network overhead**: HTTP headers, TCP/IP overhead
7. **Not planning for growth**: Consider 2-3 years of growth

### Pro Tips

1. **Start with powers of 10**: Use 10, 100, 1K, 10K, 100K, 1M, etc.
2. **Break down complex calculations**: Divide into smaller, manageable parts
3. **Validate with sanity checks**: Does the final number make sense?
4. **Consider bottlenecks**: Identify the limiting factor (CPU, memory, I/O, network)
5. **Use the 80-20 rule**: 80% of traffic comes from 20% of users
6. **Think in orders of magnitude**: Being off by 2-3x is usually acceptable
7. **Show your work**: Explain your assumptions and reasoning

### Quick Conversion Tricks

-   **1 million seconds** ≈ 11.5 days
-   **1 billion seconds** ≈ 31.7 years
-   **Rule of 72**: To find doubling time, divide 72 by growth rate percentage
-   **Daily to QPS**: Divide by 100K (86,400 ≈ 100K for rough estimates)
-   **Monthly to daily**: Divide by 30
-   **Yearly to daily**: Divide by 365 (or 400 for easier calculation)

## Practice Problems {#practice-problems}

### Problem 1: Video Streaming Platform

Design calculations for a video platform with:

-   50M daily active users
-   Average 2 hours of video watched per user per day
-   Average video bitrate: 2 Mbps
-   Video storage: multiple quality levels

**Your task**: Calculate bandwidth requirements and storage needs.

### Problem 2: Ride-Sharing Service

Calculate for a ride-sharing app with:

-   10M rides per day globally
-   Each ride generates 100 location updates
-   Each update is 200 bytes
-   Data retained for 1 year for analytics

**Your task**: Calculate storage and real-time processing requirements.

### Problem 3: E-commerce Platform

Estimate for an e-commerce site with:

-   100M registered users
-   20M daily active users
-   Peak traffic during sales: 10x normal
-   Average 5 page views per session
-   Average page size: 500KB

**Your task**: Calculate bandwidth and server requirements for peak load.

### Problem 4: Notification System

Design calculations for a push notification system:

-   1B mobile app users
-   50% daily active users
-   Average 10 notifications per active user per day
-   Each notification: 1KB payload
-   Delivery success rate: 95%

**Your task**: Calculate QPS and infrastructure needs.

## Final Recommendations

1. **Practice regularly**: Work through different scenarios to build intuition
2. **Learn key numbers**: Memorize the essential latency and throughput numbers
3. **Think systematically**: Follow the same approach for each problem
4. **Validate assumptions**: Always check if your estimates make sense
5. **Consider trade-offs**: Understand the implications of your calculations on system design
6. **Stay updated**: Technology changes, so refresh your baseline numbers periodically

Remember, the goal isn't perfect accuracy but demonstrating structured thinking and validating design decisions with reasonable estimates. Being within an order of magnitude is usually sufficient for system design interviews.
