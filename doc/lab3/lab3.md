Goals:

1. You goal is to replicate the gateway so that there are two replicas 

2. At start-up time, design a technique to
associate each sensor or device with either replica such that the
number of devices / sensors communicating with either replica is
roughly equal.
  * Nodes multicast registration requests to the gateways.
  * Gateways run a leader election algorithm.
    * The leader responds to registration requests and assigns nodes to gateways.

3. The gateway 
replicas implement a consistency technique to ensure that their states (e.g., 
database states) are synchronized.
  * We need to choose a consistency guarantee.
    * Entry and release for auto home/away.
    * Release for registration.
    * Gateway coming up pulls from other.
    * Otherwise weak consistency with a stale threshold (keep track of timestamp of latest sync)

4. Implement a cache in the front-end tier to enhance performance of the
gateway.
  * Implement a write-through cache.

5. Extend your cache consistency technique to handle
replicated data items in the cache as well as that in the database replicas. 
  * Write-through cache will take care of consistency.

6. Make your gateway fault
tolerant. It is sufficient to handle crash faults.
  * Use "I am alive" heartbeats.

7. Your failure recovery method
needs to inform the sensors / devices of the failure and have them 
reconfigure
themselves to communicate with the new replica for subsequent requests.
  * Each gateway will have knowledge of all other things in the system.
  * Upon failure of a gateway, the alive gateway will be the leader.
  * The alive gateway will take over servicing all nodes.
  * The alive gateway will notify all new nodes of their new gateway.
  * The alive gateway will query states as necessary in case any pushes were sent to the dead gateway and it did not have a chance to respond appropriately.

8. This lab does not need vector, logical clocks or leader election aspects 
of lab 2
 and it is fine to simply assume that clocks are synchronized and simple 
timestamps for determining event ordering.
