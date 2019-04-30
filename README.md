Rate - A transparent rate limiting service
------------------------------------------

Rate is a simple rate limiting service which can be deployed infront of a downstream service. It is intended to impose a per minute limit on the number of requests started per distinct resource.

## Design

### Constraints

1. Every unique HTTP path requested is considered a resource and should be limited accordingly. 
2. The number of resources which can be inflight in any given moment will be configurable with a default of 100.

### Considerations

1. Ideally this service would be horizontally scalable and have redundency. Since it acts as a transparent proxy, it should mitigate against becoming a single point of failure.  Which means deploying replicas of this service.
2. The number of distinct resources which could exist over time could grow without limit if not bound in some way. This could become problematic and the service should mitigate against that.
3. Ideally we should be able to measure the _negative_ impact this service has on request latency. In particular the ellapsed time from the beginning of the current minute, or the start of the request (whichever comes latest) to when the request is forwarded to the downstream service. This service will intentionaly create latency when request limits are being reached for certain requests. So negative impact on latency is expected in many cases. However, we want to measure how much time is spent resolving a request occuring within safe limits is wasted.
4. Requests should block idefinitely until resource becomes available or the client terminates it.

### Ideas

#### 1. Multiple cooperating proxy applications obtaining tokens (as keys) from [etcd](https://github.com/etcd-io/etcd).

Why?

- Distributed, fast, key-value store.
- Simple API with useful concurrency features like atomic operations on keys, micro-transactions and TTLs.

Downsides?

- requires persistence

How?

a. A count is made within etcd for all keys with the prefix of the current request `path` and the version of the response is noted (etcd keys and prefixes are versioned).

b. Given the number of prefixes is equal to or exeeds the current configured limit then sleep until the next minute.

c. Every request attempts to create a single key comprising of the `path` being requested concatenated onto a unique `requested ID` with a TTL rounding up to the next minute.

d. This insert attempt is made with the predicate that the version of the `path` prefix has not changed since it was first retrieved. This ensures the number of requests for the path has not changed since first counting the number of requestions.

e. If unsuccessful (the count has changed) then return to step a.

f. Otherwise, the key is created and the request can be performed.

g. Once the request finishes revoke the lease and delete the key if still alive.

#### 2. Multiple rate limiting service replicas which limit a proportion of the global limit based derived from the number of replicas.

Why?

- requires no shared persistence layer
- simpler strategy for managing limits in memory

How?

a. Implement simple semaphore using channel of structs (token bucket algorithm).

b. Have each request fetch or create a semaphore from a map keyed by request path. And block until a "token" (empty struct) can be claimed from the semaphore.

c. Once claimed perform the request and return the token once the request is finished or we reach the next interval.

d. deploy n services behind a load balancer using a round robin strategy.

e. configure each instance to limit number of requests to globally configured limit (default 100) / number of replicas.

Downsides?

- Complexity in configuration and reacting to changes like downtime in one rate limiter.

Stretch Goals

a. Implement an expiration mechanism for keys which are not being fetched. Perhaps using an LFU or LRU structure over a map? 
