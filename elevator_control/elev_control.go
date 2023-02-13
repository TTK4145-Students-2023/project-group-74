
//neccecary reocurring functions 
struct elev_output(
	executed orders vector
	aval Availability
	optional currentfloor
)
struct Availability(
	Avaliable =0,
	Busy
)

func OpenDoor{
	open door
	while opendoor timer<3
		if poll_obstuction ==1
			opendoor_timer=0
		else wait 0.1s sec
	close door
}


func decide_targetfloor{
if targetfloor=currentfloor	
	targetfloor<-cab call
else 
	targetfloor_queue <- targetfloor
	targetfloor <- cab call
}


/////////////

input: 
next_order (from master or cab call)
Hall-call matrix (recived from master) 


internal:

action master call
	targetfloor=master_Call// maybe or maybe not
	drive
	while driving towards targetfloor 
		check if currentfloor && current_direction in hMatrix
			DoorOpen()
			output elev_output
			next_action -> drive_cab_call

	if current_floor=targetfloor	
		output elev_output
		next_action -> drive_cab_call

action drive_cab_call
	decide_targetfloor()
	while driving	
		check if currentfloor && current_direction in hMatrix
			stop, dooropenfunc
			output elev_output
			next_action -> drive_cab_call

		if current_floor=targetfloor	
			output elev_output
			next_action -> idle


action idle
	motor stop















































