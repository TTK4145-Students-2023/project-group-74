# Elevator project - TTK4145

## Table of contents
* [General info](#general-info)
* [Module description](#module-description)
* [Setup](#setup)

## General info
The Elevator Project requires creating software that can control multiple elevators operating in parallel across multiple floors. The main requirements include: 
* Ensuring that no calls are lost
* Handling failure states that prevent communication between elevators or servicing of calls
* Providing sensible and efficient behavior for each elevator. 

Additionally, the system should distribute calls across elevators in a way that maximizes efficiency. 

## Module description
* ### Order Assigner 
Whenever new data enters the system, such as a new request or an update on an elevator's state, the order assigner redistributes all hall requests. This ensures that requests are assigned to different elevators at different times, promoting efficient use of the system. To maintain consistency across all elevators, a single ```MASTER``` elevator calculates the distribution using a cost function executable downloaded from the course ```Project resources```. The cost function takes into account the elevator's behavior (```idle```, ```moving```, or ```door_open```), current ```floor```, ```direction``` of travel, and current ```cab``` requests. The output of the algorithm provides updated hall requests for each elevator in the form of a list of pairs of Boolean values, indicating whether there is an ```hall_up``` or ```hall_down``` request at each floor.

* ### Network 
The system employs both ```MASTER/SLAVE``` and ```PEER-TO-PEER``` functionalities. The ```MASTER/SLAVE``` function ensures that only one elevator is in charge and instructs the others on what tasks to perform, specifically the distribution of orders among the elevators. On the other hand, the ```PEER-TO-PEER``` function is utilized to share "elevator-state" information among the elevators. This enables all elevators to retain the other elevators' data in case an elevator returns to the network after being offline for any reason.

* ### Elevator Control 
* ### Local Types 
	
## Setup
To run this project, install it locally using npm:

```
$ cd ../lorem
$ npm install
$ npm start
```



Overview of the task 

Description of the modules

Other code used

KLADD beskrivelse av moduler: 
/*
This is a Go package called elev_control that provides functionality for running an elevator.
The package contains a function called RunElevator which handels all interaction with one elevator, in all states the elevator can exist in.
*/
/*
OVERVIEW:
RunElevator runs the elevator and processes the events from the provided channels.
It starts by initializing the elevator to the first floor and setting the initial direction to up.
Then it listens to the events on the provided channels, and updates the elevator state and direction
accordingly. It also sends the state updates to the provided state channel. If the stop channel is
closed, the function stops the elevator and exits.
The function first initializes the elevator and waits for a new floor signal. It then sets up some matrices and timer for handling orders and updates.
The elevator continuously listens for new orders and floor signals.

When a new order is received, the function updates the matrices and sets the elevator in motion if it is idle.
	If the elevator is moving, it continues moving in the same direction.
	If the elevator is in the door open state, it waits for a timer to expire before closing the doors and moving to the next floor.

When a new floor signal is received, the function updates the elevator's current floor and checks if it has reached a floor with a cab call.
	If there is a cab call at the current floor, the function stops the elevator, opens the doors and waits for a timer to expire before updating the order lights and setting the elevator in motion again.
	If there is no cab call, the elevator continues in its current direction.

ARGUMENTS:
 - myIP: a string that contains the elevators IP adress
 - TxElevInfoChan: a channel to send elevator information to the OrderAssigner
 - RxElevInfoChan
 - TxNewHallRequestChan
 - RxNewHallRequestChan
 - TxFinishedHallOrderChan
 - RxFinishedHallOrderChan
 - RxNewOrdersChan
 - TxP2PElevInfoChan
 - RxP2PElevInfoChan
 - NewFloorChan
 - ObstructionChan

 RETURNS:
 - None
*/
