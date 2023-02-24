package elev_control

import "elev_control/elevio"

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

////////////////////////utlevert kode
	NumFloors := 4
	NumBtn := 3

	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Stop
	elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	//matrise som lagrer bestilligene og logikk som bestemmer hvem som skal behandles.

	for {
		select {
		case a := <-drv_buttons:
			order_matrix[a.Button][a.Floor] = 1
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			if a == numFloors-1 {
				d = elevio.MD_Down
			} else if a == 0 {
				d = elevio.MD_Up
			}
			elevio.SetMotorDirection(d)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}















































