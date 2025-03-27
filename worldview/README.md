# Worldview Module
The worldview module manages the shared elevator state across all peers. It collects elevator status data, synchronizes it with other peers, and ensures consistency despite network delays or failures. Each peer continuously updates its local elevator status, an the peers exchange their worldviews to maintain a consistent state.

The worldview module manages the shared elevator information across all peers. It collects elevator information from all peers and stores it as its own worldview. The module is an important part of the overall synchronization between peers and ensures consistency despite network delays or failures.

`worldview.go:` 
Contains most of the types, functions and variables necessary to keep track of every peer's worldview. A `WorldView`-type variable holds the information a peer has, about itself and every other peer currently on the network. It is this type that is broadcasted on the network, functioning both as heartbeat and information-channel.
The global variable `WorldViewMap` holds the worldview of every connected peer as a map indexed with the peers' IDs as keys. It is this map that is compared for all peers by the `syncing`-module.


`conversion.go:` 
Contains usefull functions for converting variables of different types to make a smooth transmission of information mainly between the `elev_algo`-, `worldview`- and `HRA`-modules


