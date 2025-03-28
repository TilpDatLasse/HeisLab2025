HRA module
================================
The HRA module is responsible for distributing hall requests among available elevators in the P2P network. It ensures an efficient deterministic allocation of requests while maintaining system synchronization.
After requesting synchronization, it waits for the synchronized worldview. The worldview is fed to the hall request assigner and its output sent to the `elev_algo`-module.



