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
The program is written in ```GO```, because it is a compiled language with a strong type system, which makes it efficient and performant. Additionally, GO has built-in support for concurrency and parallelism, which can be useful for controlling multiple elevators in a building. GO also has a large standard library that includes networking and serialization capabilities, making it easy to communicate with other components of the elevator system. Finally, GO's simple and concise syntax can make it easier to write and maintain code, especially for large projects like an elevator system.

Multiple of the projects modules have auxiliary functionality and are defined inside the same folder. 

### Order Assigner 
Whenever new data enters the system, such as a new request or an update on an elevator's state, the order assigner redistributes all hall requests. This ensures that requests are assigned to different elevators at different times, promoting efficient use of the system. To maintain consistency across all elevators, a single ```MASTER``` elevator calculates the distribution using a cost function executable downloaded from the course ```Project resources```. The cost function takes into account the elevator's behavior (```idle```, ```moving```, or ```door_open```), current ```floor```, ```direction``` of travel, and current ```cab``` requests. The output of the algorithm provides updated hall requests for each elevator in the form of a list of pairs of Boolean values, indicating whether there is an ```hall_up``` or ```hall_down``` request at each floor.

The code and explanation for the hall_request_assigner is given here: https://github.com/TTK4145/Project-resources/tree/master/cost_fns/hall_request_assigner.

### Network 
The network is ```TCP``` and employs both ```MASTER/SLAVE``` and ```PEER-TO-PEER``` functionalities. The ```MASTER/SLAVE``` function ensures that only one elevator is in charge and instructs the others on what tasks to perform, specifically the distribution of orders among the elevators. On the other hand, the ```PEER-TO-PEER``` function is utilized to share "elevator-state" information among the elevators. This enables all elevators to retain the other elevators' data in case an elevator returns to the network after being offline for any reason.

### Elevator Control
This module is responsible for running an elevator control system. It communicates with other elevators the control system to manage requests and operate the elevator. The main function runs the elevator control logic in a loop, which processes new orders and updates the elevator state accordingly.

### Local Types 
This module defines a number of constants, type definitions, and validation functions for an elevator control system. The constants include the number of ```buttons```, ```floors```, and orders, as well as various time intervals and update intervals. The type definitions include button types, button information, elevator states, motor directions, elevator information, and input/output structures. The validation functions ensure that inputs to the system are valid.	

## Setup
To run this project, run the following command in an open terminal on 3 of the computers at Sanntidlaben. Requires an active ```elevatorserver```

```
$ go run main.go
```


