HRA module
================================
The HRA module is responsible for distributing hall requests among available elevators in the P2P network. It ensures an efficient deterministic allocation of requests while maintaining system synchronization.

Key components of the module are `HRAMain`,`sendToElev` and `hallToBool`.

`HRAMain` is the main loop for the module, which syncs elevator states, formats data, runs HRA, and distributes assignments. 

`sendToElev` sends updated hall requests to the correct elevator.

`hallToBool` converts confirmation states to boolean values for HRA processing.

