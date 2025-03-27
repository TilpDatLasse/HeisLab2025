Elev_algo Module
================
The elevator algorithm module controls a single elevator over TCP, managing movement, state transitions, and user input. It comprises four submodules: elevator_io, fsm, requests and timer. 

The elevator algorithm is based on preferring to continue in the direction of travel, as long as there are any requests in that direction. 
We implement thhis with three functions, ChooseDirection, ShouldStop and ClearAtCurrentFloor in [requests.go](requests/requests.go).

Finite State Machine (FSM)
-------------
The FSM manages the elevator's state transitions between Idle, Moving, and DoorOpen. It ensures that the elevator moves when there are active requests, stops when necessary and the door opens and closes appropriately.

Requests
-----
 ClearAtCurrentFloor

Timer
-----

Elevator_io
----------

This module interfaces with hardware to send movement commands, read floor sensors, and detect button presses. It is based on code from the [driver-go] (https://github.com/TTK4145/driver-go.git) repository.


The basic elevator algorithm
-------------

The elevator algorithm is based on preferring to continue in the direction of travel, as long as there are any requests in that direction. We implement this algorithm with three functions:
 - Choose direction:
   - Continue in the current direction of travel if there are any further requests in that direction
   - Otherwise, change direction if there are requests in the opposite direction
   - Otherwise, stop and become idle
 - Should stop:
   - Stop if there are passengers that want to get off at this floor
   - Stop if there is a request in the direction of travel at this floor 
   - Stop if there are no further requests in this direction
 - Clear requests at floor:  
   This function comes in two variants. We can either assume that anyone waiting for the elevator gets on the elevator regardless of which direction it is traveling in, or that they only get on the elevator if the elevator is going to travel in the direction the passenger desires. (Most people would expect the first behaviour, but there are elevators that only clear the requests "in the direction of travel". I believe the one outside EL6 behaves like this.)
   - Always clear the request for getting off the elevator and the request for entering the elevator in the direction of travel
   - Either:
     - A: Always clear the request for entering the elevator in the opposite direction
     - B: Clear the request in the opposite direction if there are no further requests in the direction of travel
     
The implementations of these three functions are found in [requests.go](requests/requests.go).
