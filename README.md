Rate - A rate limiting service
------------------------------------------

Rate is a simple rate limiting service which can be deployed infront of a downstream service.
It is intended to impose a per minute limit on the number of requests started for each distinct resource.

## Design

### Constraints

1. Every unique HTTP path requested is considered a resource and should be limited accordingly.
2. The number of resources which can be inflight in any given moment will be configurable with a default of 100. Though this may vary slightly under certain conditions (re-balancing due to scale up / down).

### Considerations

1. Ideally this service would be horizontally scalable and have redundency. Since it acts as a transparent proxy, it should mitigate against becoming a single point of failure.  Which means deploying replicas of this service.
2. The number of distinct resources which could exist over time could grow without limit if not bound in some way. This could become problematic and the service should mitigate against that.
3. Ideally we should be able to measure the _negative_ impact this service has on request latency. In particular the ellapsed time from the beginning of the current minute, or the start of the request (whichever comes latest) to when the request is forwarded to the downstream service. This service will intentionaly create latency when request limits are being reached for certain requests. So negative impact on latency is expected in many cases. However, we want to measure how much time is spent resolving a request occuring within safe limits is wasted.
4. Requests should block idefinitely until resource becomes available or the client terminates it.

### Ideas

#### 1. Multiple cooperating proxy rate-limiters obtaining tokens (as keys) from [etcd](https://github.com/etcd-io/etcd).

##### Why?

Makes configuration, scale and reaction to failures automatic for the proxy service. Cooperation with a shared bucket implemented within etcd means a lot of coordination happens using etcd primitives like TTLs on keys and micro-transactions. With this we can maintain a hard global upper limit on requests for resources.

##### How?

1. A count is made within etcd for all keys with the prefix of the current request `path` and the version of the response is noted (etcd keys and prefixes are versioned).
2. Given the number of keys with prefix `path` is equal to or exeeds the current configured limit then sleep until the next minute.
3. Every request attempts to create a single key comprising of the `path` being requested concatenated onto a unique `requested ID` with a TTL rounding up to the next minute.
4. This insert attempt is made with the predicate that the version of the `path` prefix has not changed since it was first retrieved. This ensures the number of requests for the path has not changed since counting the number of requestions.
5. If unsuccessful (the count has changed) then return to step 1.
6. Otherwise, the key is created and the request can be performed.
7. Once the request finishes revoke the lease and delete the key if still alive.

##### Downsides and Complexities

- Requires deploying and managing an etcd cluster.
- Adds complexity in the form of a cooperation algorithm using etcd primitives.
- A strategy for iterating on this cooperation algorithm would ideally need to be planned. How does we make changes to this? For example, do we need to version the keyspace?
- Adds overhead in the form of finding consensus between replicas for each request. Hard to estimate how much at this stage though easily measured.

#### 2. Round-robin load balanced set of rate limiters which each enforce `global limit / number of replicas` locally.

##### Why?

- Requires no shared persistence layer.
- Simpler strategy for managing limits in process.
- Requires no consensus with external actors, likely leading to more performance (hypothesis).

##### How?

1. Implement simple semaphore using channel of structs (token bucket algorithm).
2. Have each request fetch or create a semaphore from a map keyed by request path. And block until a "token" (empty struct) can be claimed from the semaphore.
3. Once claimed perform the request and return the token once the request is finished or we reach the next interval.
4. deploy n services behind a load balancer using a round robin strategy.
5. configure each instance to limit number of requests to globally configured limit (default 100) / number of replicas.

##### Downsides / Complexities

Scaling and reacting to failure won't come for free. External mechanisms are required to identify and react to these situations, in order to re-balance limits and apply new configuration.

To support this, complexity may need to be introduced into the deployment strategy or in the token leasing implementation. For example, exposing a configuration endpoint on the rate limiters to change the limit and react to this inflight. In this situation we might want to loosen the constraints on the global inflight limit (e.g. sometimes the limit globally might let through just over 100 requests) in order to simplify strategy and find eventual consistency.

##### Stretch Goals

1. Implement an expiration mechanism for keys which are not being fetched. Perhaps using an LFU or LRU structure over a map?
