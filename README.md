This code is a triple threat, quite literally! It runs not one, not two, but THREE elevators. That's right, three times the fun, three times the excitement, and three times the potential for awkward elevator small talk.

But fear not, this code is no joke. It's been rigorously tested to ensure smooth operation and minimal wait times. You'll never have to worry about being stuck in a cramped elevator with that one person who insists on sharing their life story (we all know that person).

So come on, hop on board and let this code take you to new heights (literally). And who knows, maybe you'll even make a new elevator buddy or two. Just remember to keep the conversations light and fluffy, like the clouds you'll be soaring above. Happy elevator riding!

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
