Network module 
==========================================

Network module responsible for UDP communication between the peers. The local peer's worldveiw is broadcasted and the recieved worldviews are passed on back to the worldview module.
The submodules `bcast`, `conn` and `peers` are modified from handout code for the project found [here](https://github.com/TTK4145/Network-go.git).

`bcast.go:` Handles the actual transmission and reception on the udp-network.

`peers.go:` Keeps track of online peers, using the transmitted worldviews as a heartbeat-mechanism.
