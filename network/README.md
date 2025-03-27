Network module 
==========================================

Network module based on modified handout code for the project found at : linklink

Channel-in/channel-out pairs of (almost) any custom or built-in data type can be supplied to a pair of transmitter/receiver functions. 
Data sent to the transmitter function is automatically serialized and broadcast on the specified port. 
Any messages received on the receiver's port are de-serialized (as long as they match any of the receiver's supplied channel data types) and sent on the corresponding channel. 
See [bcast.Transmitter and bcast.Receiver](bcast/bcast.go). 

Peers on the local network can be detected by supplying your own ID to a transmitter and receiving peer updates (new, current, and lost peers) from the receiver. See [peers.Transmitter and peers.Receiver](peers/peers.go).
