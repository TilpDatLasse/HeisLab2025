HRA module
================================
The HRA module is responsible for distributing hall requests among available elevators in the P2P network. It ensures an efficient deterministic allocation of requests while maintaining system synchronization.

## Key Components
### `HRAMain`
- Main loop for the module.
- Syncs elevator states, formats data, runs HRA, and distributes assignments.

### `sendToElev(output, ch_elevatorQueue, ID)`
- Sends updated hall requests to the correct elevator.

### `hallToBool(hallReqList)`
- Converts confirmation states to boolean values for HRA processing.

## Dependencies
- **Elevator Module** (`elev_algo/elevator_io`)
- **WorldView Module** (`worldview`)
- **HRA Executable** (platform-specific hall request assigner)
