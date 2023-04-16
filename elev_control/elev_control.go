package elev_control

import (
	"fmt"
	"project-group-74/elev_control/elevio"
	"project-group-74/localTypes"
	"project-group-74/network"
	"time"
)

// ----- MAIN FUNCTION (ELEVATOR CONTROL) ------ //
func RunElevator(
	myIP string,
	TxElevInfoChan chan<- localTypes.LOCAL_ELEVATOR_INFO,
	RxElevInfoChan chan<- localTypes.LOCAL_ELEVATOR_INFO,
	TxNewHallRequestChan chan<- localTypes.BUTTON_INFO,
	RxNewHallRequestChan chan<- localTypes.BUTTON_INFO,
	TxFinishedHallOrderChan chan<- localTypes.BUTTON_INFO,
	RxFinishedHallOrderChan chan<- localTypes.BUTTON_INFO,
	RxNewOrdersChan <-chan map[string]localTypes.HMATRIX,
	TxP2PElevInfoChan chan<- localTypes.P2P_ELEV_INFO,
	RxP2PElevInfoChan <-chan localTypes.P2P_ELEV_INFO,
	NewFloorChan <-chan int,
	NewBtnPressChan <-chan localTypes.BUTTON_INFO,
	ObstructionChan <-chan bool) {

	MyElev :=
		localTypes.LOCAL_ELEVATOR_INFO{
			State:     localTypes.Idle,
			Floor:     -1,
			Direction: localTypes.DIR_stop,
			CabCalls:  [localTypes.NUM_FLOORS]bool{},
			ElevID:    myIP,
		}

	initializing := true

	for initializing {
		select {
		case MyElev.Floor = <-NewFloorChan:
			elevio.SetMotorDirection(localTypes.DIR_stop)
			elevio.SetDoorOpenLamp(false)
			network.SendlocalElevInfo(MyElev, RxElevInfoChan, TxElevInfoChan)
			initializing = false
			fmt.Printf("Initializing finished!\n")
		default:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(localTypes.DIR_down)
			time.Sleep(80 * time.Millisecond)
		}
	}
	fmt.Printf("My Elev: %+v\n", MyElev)

	var MyOrders localTypes.HMATRIX
	var CombinedHMatrix localTypes.HMATRIX
	AllElevs := make(localTypes.P2P_ELEV_INFO, 0)

	elevio.UpdateOrderLights(MyElev, CombinedHMatrix)

	var dooropentimer *time.Timer
	dooropentimer = time.NewTimer(time.Second * 1000)
	dooropentimer.Stop()

	for {
		select {
		case newOrder := <-RxNewOrdersChan:
			if MyOrders != newOrder[MyElev.ElevID] {
				MyOrders = newOrder[MyElev.ElevID]
			}
			fmt.Printf("Myorders %+v\n", MyOrders)
			CombinedHMatrix = elevio.AddNewOrdersToHMatrix(newOrder)
			elevio.UpdateOrderLights(MyElev, CombinedHMatrix)

			switch MyElev.State {
			case localTypes.Door_open:
			case localTypes.Moving:
			case localTypes.Idle:
				if elevio.IsOrderAtFloor(MyElev, MyOrders) {
					MyElev.CabCalls[MyElev.Floor] = false
					if MyOrders[MyElev.Floor][localTypes.Button_hall_up] {
						network.SendButtonInfo(MyElev, localTypes.Button_hall_up, RxFinishedHallOrderChan, TxFinishedHallOrderChan)

					} else if MyOrders[MyElev.Floor][localTypes.Button_hall_down] {
						network.SendButtonInfo(MyElev, localTypes.Button_hall_down, RxFinishedHallOrderChan, TxFinishedHallOrderChan)

					}
					MyElev.State = localTypes.Door_open
					elevio.SetDoorOpenLamp(true)
					dooropentimer = time.NewTimer(localTypes.OPEN_DOOR_TIME_sek * time.Second)
					network.SendlocalElevInfo(MyElev, RxElevInfoChan, TxElevInfoChan)

				} else {
					newDir, newState := elevio.FindDirection(MyElev, MyOrders)
					MyElev.Direction, MyElev.State = newDir, newState
					elevio.SetMotorDirection(newDir)
					network.SendlocalElevInfo(MyElev, RxElevInfoChan, TxElevInfoChan)
				}
			}
		case newFloor := <-NewFloorChan:
			if MyElev.Floor != newFloor {
				MyElev.Floor = newFloor
				elevio.SetFloorIndicator(MyElev.Floor)
			}
			var changed bool
			var nextdir localTypes.MOTOR_DIR
			var nextstate localTypes.ELEVATOR_STATE
			if MyElev.CabCalls[MyElev.Floor] {
				MyElev.CabCalls[MyElev.Floor] = false
				elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
				nextdir, nextstate = localTypes.DIR_stop, localTypes.Door_open
				changed = true
			}
			switch MyElev.Direction {
			case localTypes.DIR_stop:
			case localTypes.DIR_up:
				switch {
				case MyOrders[MyElev.Floor][localTypes.Button_hall_up]:
					nextdir, nextstate = localTypes.DIR_stop, localTypes.Door_open
					changed = true
					network.SendButtonInfo(MyElev, localTypes.Button_hall_up, RxFinishedHallOrderChan, TxFinishedHallOrderChan)

				case elevio.Requests_above(MyElev, MyOrders):

				case MyOrders[MyElev.Floor][localTypes.Button_hall_down]:
					nextdir, nextstate = localTypes.DIR_stop, localTypes.Door_open
					changed = true
					network.SendButtonInfo(MyElev, localTypes.Button_hall_down, RxFinishedHallOrderChan, TxFinishedHallOrderChan)

				default:
					MyElev.State, MyElev.Direction = localTypes.Idle, localTypes.DIR_stop
					elevio.SetMotorDirection(MyElev.Direction)
				}

			case localTypes.DIR_down:
				switch {
				case MyOrders[MyElev.Floor][localTypes.Button_hall_down]:
					nextdir, nextstate = localTypes.DIR_stop, localTypes.Door_open
					changed = true
					network.SendButtonInfo(MyElev, localTypes.Button_hall_down, RxFinishedHallOrderChan, TxFinishedHallOrderChan)

				case elevio.Requests_below(MyElev, MyOrders):

				case MyOrders[MyElev.Floor][localTypes.Button_hall_up]:
					nextdir, nextstate = localTypes.DIR_stop, localTypes.Door_open
					changed = true
					network.SendButtonInfo(MyElev, localTypes.Button_hall_up, RxFinishedHallOrderChan, TxFinishedHallOrderChan)

				default:
					MyElev.State, MyElev.Direction = localTypes.Idle, localTypes.DIR_stop
					elevio.SetMotorDirection(MyElev.Direction)
				}
			}
			if changed {
				MyElev.Direction, MyElev.State = nextdir, nextstate
				elevio.SetMotorDirection(nextdir)
				if nextstate == localTypes.Door_open {
					elevio.SetDoorOpenLamp(true)
					dooropentimer = time.NewTimer(localTypes.OPEN_DOOR_TIME_sek * time.Second)
				}
			}
			network.SendlocalElevInfo(MyElev, RxElevInfoChan, TxElevInfoChan)

		case newBtnPress := <-NewBtnPressChan:
			switch newBtnPress.Button {
			case localTypes.Button_Cab:
				MyElev.CabCalls = elevio.AddOneNewOrderBtn(newBtnPress, MyElev)
				elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
				switch MyElev.State {
				case localTypes.Moving:
				case localTypes.Door_open:
					if newBtnPress.Floor == MyElev.Floor {
						MyElev.CabCalls = elevio.RemoveOneOrderBtn(newBtnPress, MyElev)
						elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
						dooropentimer.Reset(localTypes.OPEN_DOOR_TIME_sek * time.Second)
					}

				case localTypes.Idle:
					MyElev.Direction, MyElev.State = elevio.FindDirection(MyElev, MyOrders)
					elevio.SetMotorDirection(MyElev.Direction)
					if MyElev.State == localTypes.Door_open {
						elevio.SetDoorOpenLamp(true)
						MyElev.CabCalls[newBtnPress.Floor] = false
						elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
						dooropentimer = time.NewTimer(localTypes.OPEN_DOOR_TIME_sek * time.Second)
					}
					network.SendlocalElevInfo(MyElev, RxElevInfoChan, TxElevInfoChan)
				}
			case localTypes.Button_hall_up:
				if !elevio.IsHOrderActive(newBtnPress, CombinedHMatrix) {
					network.SendButtonPress(MyElev, newBtnPress, RxNewHallRequestChan, TxNewHallRequestChan)
				}

			case localTypes.Button_hall_down:
				if !elevio.IsHOrderActive(newBtnPress, CombinedHMatrix) {
					network.SendButtonPress(MyElev, newBtnPress, RxNewHallRequestChan, TxNewHallRequestChan)
				}
			}

		case obstruction := <-ObstructionChan:
			if obstruction {
				dooropentimer.Stop()
			}
			if MyElev.State == localTypes.Door_open {
				if !obstruction {
					dooropentimer = time.NewTimer(localTypes.OPEN_DOOR_TIME_sek * time.Second)
				}

			}

		case <-dooropentimer.C:
			elevio.SetDoorOpenLamp(false)
			MyElev.Direction, MyElev.State = elevio.FindDirectionNotHere(MyElev, MyOrders)
			elevio.SetMotorDirection(MyElev.Direction)
			if MyElev.State == localTypes.Door_open {
				elevio.SetDoorOpenLamp(true)
				dooropentimer = time.NewTimer(localTypes.OPEN_DOOR_TIME_sek * time.Second)
			}
			network.SendlocalElevInfo(MyElev, RxElevInfoChan, TxElevInfoChan)

		case NewAllElevs := <-RxP2PElevInfoChan:
			AllElevs = elevio.AddNewAllElevs(AllElevs, NewAllElevs)
			AllElevs = elevio.UpdateLocalInAllElevs(MyElev, AllElevs)
			fmt.Printf("\n\n\n\nrecieved this wonderful piece of information! \n\n\n")

			TxP2PElevInfoChan <- AllElevs
		}
	}

}
