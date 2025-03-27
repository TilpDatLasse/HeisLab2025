# Syncing Module  

## 📌 Overview  
The `syncing` module ensures **synchronization of elevator state data** across all peers in the peer-to-peer network. This module is crucial for **maintaining a consistent and accurate worldview** among all elevators, preventing desynchronization issues caused by network delays or failures.  

When a peer detects an inconsistency or requests a sync, this module **compares and updates the worldview**, ensuring all peers have the same elevator status information.  

---

## ⚙️ **How It Works**  
1. **Receives a sync request** from the Hall Request Assigner (HRA) or another peer.  
2. **Locks worldview data** to prevent inconsistencies during the sync process.  
3. **Compares and updates elevator state data** across peers using `worldview.CompareAndUpdateInfoMap()`.  
4. **Confirms synchronization** when all peers have consistent worldviews.  
5. **Releases the lock**, completing the sync process.  

---

## 📦 **Module Responsibilities**  
| Function | Description |
|----------|-------------|
| `SyncingMain()` | Listens for sync requests and initiates synchronization when required. |
| `Sync()` | Compares worldviews across peers and updates state data if inconsistencies are found. |
| `syncDone()` | Notifies the system when synchronization is complete. |
| `AllWorldViewsEqual()` | Checks if all peers have identical worldview data. |

---

## 🔧 **How to Use**  
The module runs automatically within the peer-to-peer system. However, if you need to manually trigger a sync, send `true` to the `ShouldSync` channel.  

Example usage:  
```go
syncChans.ShouldSync <- true // Request a sync
