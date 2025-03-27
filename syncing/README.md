# Syncing Module  

The syncing module ensures synchronization of elevator state data across all peers in the network. When a sync request is received from the Hall Request Assigner (HRA), or another peer, this module updates elevator state data across all peers to prevent inconsistencies caused by network delays or failures.

When a peer detects an inconsistency or requests a sync, this module compares and updates the worldview, ensuring all peers have the same elevator status information.  
