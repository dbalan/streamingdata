# Guarantees

1. gRPC stream, and the underlying tcp promises that the messages are received in order.
2. When a stream is closed, client know of it  (again from gRPC streams)
3. In case of stateless request client can never loose the current state of the request, and we only need the last element and remaining number of elements to restart the stream (which client sents as request argument)
4. In case of statefull request, we need to keep track of the overall hash of the response stream - which is computed again when a client reconnect, we ensure this is deterministic by using a deterministic RNG and re-seeding it with same exact seed (this is stored as part of state). We also keep track of number of elements already sent, so as to reduce the state in client for statefull request. Again we are using (1,2) to ensure that what server knows are sent are received by the client.

# Limits
1. Number of tcp connections that can be achieved.
2. Collision in clientID could potentially screw up things.
3. a lot of statefull re-connects will force recomputing the hash, thus increasing CPU resources.

