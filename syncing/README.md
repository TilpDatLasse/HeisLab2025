# Syncing Module  

The syncing module ensures synchronization of elevator information across all peers in the network. When a sync request is received from the Hall Request Assigner (HRA), this module ensures every peer holds the same worldview before sending it to the HRA. This is crucial 

When a peer detects an inconsistency or requests a sync, this module compares and updates the worldview, ensuring all peers have the same elevator status information.  
