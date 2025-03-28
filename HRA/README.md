HRA module
================================
The HRA module is responsible for distributing hall requests among available elevators in the P2P network. It ensures an efficient deterministic allocation of requests while maintaining system synchronization.

Key components of the module are `HRAMain`,`sendToElev(output, ch_elevatorQueue, ID)` and `hallToBool(hallReqList)`.

`HRAMain` is the main loop for the module, which syncs elevator states, formats data, runs HRA, and distributes assignments. 

`sendToElev(output, ch_elevatorQueue, ID)` sends updated hall requests to the correct elevator.

`hallToBool(hallReqList)` converts confirmation states to boolean values for HRA processing.

## Dependencies
- **Elevator Module** (`elev_algo/elevator_io`)
- **WorldView Module** (`worldview`)
- **HRA Executable** (platform-specific hall request assigner)
