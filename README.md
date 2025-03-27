
Elevator project TTK4145 Spring 2025
==========================================


This is a peer-to-peer elevator system. The code will launch a single peer to the elevator network. We chose a peer-to-peer system as we found this the easiest, as having a master-slave system would be harder to control when packetloss and network connection loss occurs.

The system will operate a single elevator if launched on its own. When multiple peers are launched, the peers will communicate over udp to ensure service of elevator orders. The system should function when packetloss occurs, as well as when sudden failures in the system happens. The full functionality specifications can be found [here](https://github.com/TTK4145/Project.git).



Modules
--------

The system consists of several modules, responsible for different operations within the program. Here is a short list of each modules responsibilities:

elev_algo: Runs a single assigned elevator over TCP communication. Responsible for finite state machine of the elevator and bulletproof excecution of elevatormovements and user input and output.

HRA: Is the Hall request assigner module, which controls which of the online elevators that should serve a hallorder when multiple elevators are online.

network: Responsible for all udp communication between peers. Information transmitting and receiving are both done in this module.

syncing: 

worldview:

How to run the program
-----------------------

How to run the program:

To run the code, one will have to start either the physival elevator server or the simulator found [here](https://github.com/TTK4145/Simulator-v2.git).

When running the program in the terminal, you have the option to set a flag for the peer name (id), UDP ports (udpWVPort and udpPeersPort) and TCP Port (simPort) between the program and the server.

To set a a flag when running the code, include -FLAGIDENTIFER=value, where FLAGIDENTIFIER is the flagtaht should be set and value if the value it should be set to. If no flags are set, the code will run with the standard values:

id = one
simPort = 15657
udpWVPort = 14700
udpPeersPort = 16500

Example of how to run the code with flags:

go run main.go -id=two -simPort=11111 -udpWVPort=22222 -udpPeersPort=33333


