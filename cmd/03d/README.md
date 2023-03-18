# Challenge #3d: Efficient Broadcast, Part I

Personal results:

```plaintext
:servers {:send-count 42864,
            :recv-count 42864,
            :msg-count 42864,
            :msgs-per-op 24.162346},

:stable-latencies {0 0,
                    0.5 373,
                    0.95 491,
                    0.99 503,
                    1 505},
```

Optimizations:

1. Skip broadcasting back to the node that broadcasted to us. The sending node
   already has the message. This reduced `msgs-per-op` from ~45 to ~25.
2. Increase the backoff time. Saves us from more unncessary message operations.
   The takeaway here is that backoff duration should be tailored to expected
   latency in the network.
3. Use `--topology tree4` for the incoming network topology. Technically
   cheating here since I didn't arrange the node topology on my own.

Resources:

1. Maelstrom docs – https://github.com/jepsen-io/maelstrom/blob/main/doc/03-broadcast/02-performance.md
2. @teivah's solutions – https://github.com/teivah/gossip-glomers

These two resources provided a lot of insight that helped me with this challenge.
