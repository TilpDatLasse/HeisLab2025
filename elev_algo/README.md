
Elev_algo Module
================
The elevator algorithm module controls a single elevator over TCP, managing movement, state transitions, and user input. The module is based on a 5-state state machine (States: {INIT, IDLE, MOVE, STOP,DOOROPEN}, Events: {Button press, Arrive at floor, Timer timed out, order from HRA, }). It comprises four submodules:
  - `elevator_io` : Interfaces with hardware to send movement commands, read floor sensors and detect button presses. It is based on handout code from the [driver-go](https://github.com/TTK4145/driver-go.git) repository.
  - `fsm` : 
    - `fsm.go:` The state machine controlling the single local elevator. 
    - `failures.go:` Ensures consistency in case of failure states.
  - `requests`  : Handles requests in the right manner and helps the fsm make the right decisions.
  - `timer`: Handles timers in the whole system, including the dooropen-timer.


The basic elevator algorithm
============================

The elevator algorithm is based on preferring to continue in the direction of travel, as long as there are any requests in that direction. We implement this algorithm in the `requests`-module. The algorithm can be described as follows:
 - Choose direction:
   - Continue in the current direction of travel if there are any further requests in that direction
   - Otherwise, change direction if there are requests in the opposite direction
   - Otherwise, stop and become idle
 - Stop:
   - If there are passengers that want to get off at this floor
   - If there is a request in the direction of travel at this floor 
   - If there are no further requests in this direction
 - Clearing requests:  
   - We will assume that passengers will only enter the elevator if it is traveling in their desired direction (i.e. we only clear requests in the direction of travel) 
   - Cab requests are always cleared at the current floor






