Syncing Module
================

The syncing module ensures synchronization of elevator information across all peers on the network. When a sync request is received from the Hall Request Assigner (HRA), this module ensures every peer holds the same worldview before sending it to the HRA. This is crucial, since different inputs would provide the seperate peers with different combinations of assigned orders.

An important function is `AllWorldViewsEqual()` which compares all the worldviews in the global `WorldViewMap`. The map is continuously updated with information recieved on the network in the worldview module. Since the information a peer sends is `Locked` during synchronization, all worldviews will eventually be the same. 
