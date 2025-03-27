# Worldview Module

The worldview module manages the shared elevator state across all peers. It collects elevator status data, synchronizes it with other peers, and ensures consistency despite network delays or failures.
Each peer continuously updates its local elevator status, an the peers exchange their worldviews to maintain a consistent state.
If inconsistencies or synchronization requests are detected, the system updates and locks elevator data accordingly.
The module ensures data is in sync before allowing new commands to be processed.

