
Elevator Project - TTK4145 Spring 2025
==========================================

This project implements a peer-to-peer elevator control system. Instead of a traditional master-slave architecture, we chose a P2P approach to increase fault tolerance and handle packet loss more effectively.

The system will operate a single elevator if launched on its own. When multiple peers are launched, the peers will communicate over UDP to ensure service of all elevator orders. The system should function when packet loss occurs, as well as when sudden failures in the system happen. The full functionality specifications for the project can be found [here](https://github.com/TTK4145/Project.git).


Features
--------
✔ Peer-to-peer communication for distributed elevator control  
✔ Fault-tolerant design that handles network failures  
✔ Hall Request Assigner (HRA) for efficient elevator dispatching    
✔ UDP for inter-peer communication and TCP for elevator control 


Modules
--------

The system consists of several modules, responsible for different operations within the program. 
Here is a short list of each module and their responsabilities:

elev_algo: Runs a single assigned elevator over TCP communication. Responsible for a finite state machine of the elevator and bulletproof excecution of elevator movements and user input and output.

HRA (Hall Request Assigner): Decides which elevator should serve a hall order when multiple elevators are online.

network: Responsible for all udp communication between peers. Information transmitting and receiving are both done in this module.

syncing: Ensures synchronization of elevator state data across all peers in the peer-to-peer network. 

worldview: Manages the shared worldview across all peers.

How to run the program:
-----------------------

To run the code, one will have to start either the physical elevator server or the elevator simulator found [here](https://github.com/TTK4145/Simulator-v2.git).

When running the program in the terminal, you have the option to set a flag for the peer name (id), UDP ports (udpWVPort and udpPeersPort) and TCP Port (simPort) between the program and the server.

To set a a flag when running the code, include -FLAGIDENTIFER=value, where FLAGIDENTIFIER is the flag that should be set and value if the value it should be set to. If no flags are set, the code will run with the standard values:

id = one
simPort = 15657
udpWVPort = 14700


Example of how to run the code with flags:

go run main.go -id=two -simPort=11111 -udpWVPort=22222


