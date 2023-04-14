package elev_control

import (
	"fmt"
	"project-group-74/elev_control/elevio"
	"project-group-74/localTypes"
	"time"
)

func RunElevator(myIP string,
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
			localTypes.SendlocalElevInfo(MyElev, RxElevInfoChan, TxElevInfoChan)
			initializing = false
			fmt.Printf("Initializing finished!\n")
		default:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(localTypes.DIR_down)
			time.Sleep(80 * time.Millisecond)
		}
	}
	fmt.Printf("My Elev: %+v\n", MyElev)

	var MyOrders localTypes.HMATRIX        // FOR DRIVING combined with myCabCalls
	var CombinedHMatrix localTypes.HMATRIX // FOR LIGHTS and reboot if you become master
	AllElevs := make(localTypes.P2P_ELEV_INFO, 0)

	elevio.UpdateOrderLights(MyElev, CombinedHMatrix)

	var dooropentimer *time.Timer
	dooropentimer = time.NewTimer(time.Second * 1000)

	for {
		select {
		// case newCabCalls := <- CabCallsChan:
		// 	MyElev.CabCalls = newCabCalls
		// 	switch MyElev.State{
		// 	case localTypes.Idle:
		// 	case localTypes.Moving:
		// 	case localTypes.Door_open:

		// 	}
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
						localTypes.SendButtonInfo(MyElev, localTypes.Button_hall_up, RxFinishedHallOrderChan, TxFinishedHallOrderChan) //change to TX

					} else if MyOrders[MyElev.Floor][localTypes.Button_hall_down] {
						localTypes.SendButtonInfo(MyElev, localTypes.Button_hall_down, RxFinishedHallOrderChan, TxFinishedHallOrderChan) //change to TX

					}
					MyElev.State = localTypes.Door_open
					elevio.SetDoorOpenLamp(true)
					fmt.Printf("Dooropen from neworder\n")
					dooropentimer = time.NewTimer(localTypes.OPEN_DOOR_TIME_sek * time.Second)
					localTypes.SendlocalElevInfo(MyElev, RxElevInfoChan, TxElevInfoChan)

				} else {
					newDir, newState := elevio.FindDirection(MyElev, MyOrders)
					MyElev.Direction, MyElev.State = newDir, newState
					elevio.SetMotorDirection(newDir)
					localTypes.SendlocalElevInfo(MyElev, RxElevInfoChan, TxElevInfoChan)
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
					localTypes.SendButtonInfo(MyElev, localTypes.Button_hall_up, RxFinishedHallOrderChan, TxFinishedHallOrderChan) //change to TX

				case elevio.Requests_above(MyElev, MyOrders):
					fmt.Printf("Drive By shooting\n")

				case MyOrders[MyElev.Floor][localTypes.Button_hall_down]:
					nextdir, nextstate = localTypes.DIR_stop, localTypes.Door_open
					changed = true
					localTypes.SendButtonInfo(MyElev, localTypes.Button_hall_down, RxFinishedHallOrderChan, TxFinishedHallOrderChan) //change to TX
				default:
					MyElev.State, MyElev.Direction = localTypes.Idle, localTypes.DIR_stop
					elevio.SetMotorDirection(MyElev.Direction)
				}
			case localTypes.DIR_down:
				switch {
				case MyOrders[MyElev.Floor][localTypes.Button_hall_down]:
					nextdir, nextstate = localTypes.DIR_stop, localTypes.Door_open
					changed = true
					localTypes.SendButtonInfo(MyElev, localTypes.Button_hall_down, RxFinishedHallOrderChan, TxFinishedHallOrderChan) //change to TX

				case elevio.Requests_below(MyElev, MyOrders):

				case MyOrders[MyElev.Floor][localTypes.Button_hall_up]:
					nextdir, nextstate = localTypes.DIR_stop, localTypes.Door_open
					changed = true
					localTypes.SendButtonInfo(MyElev, localTypes.Button_hall_up, RxFinishedHallOrderChan, TxFinishedHallOrderChan) //change to TX
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
			localTypes.SendlocalElevInfo(MyElev, RxElevInfoChan, TxElevInfoChan)

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
					newDir, newState := elevio.FindDirection(MyElev, MyOrders)
					MyElev.Direction = newDir
					MyElev.State = newState
					elevio.SetMotorDirection(newDir)
					if newState == localTypes.Door_open {
						elevio.SetDoorOpenLamp(true)
						MyElev.CabCalls[newBtnPress.Floor] = false
						elevio.UpdateOrderLights(MyElev, CombinedHMatrix)

						dooropentimer = time.NewTimer(localTypes.OPEN_DOOR_TIME_sek * time.Second)
					}
					localTypes.SendlocalElevInfo(MyElev, RxElevInfoChan, TxElevInfoChan) //Change to TX
				}
			case localTypes.Button_hall_up:
				if !elevio.IsHOrderActive(newBtnPress, CombinedHMatrix) {
					if len(localTypes.PeerList.Peers) == 0 {
						RxNewHallRequestChan <- newBtnPress
					} else {
						TxNewHallRequestChan <- newBtnPress
					}
				}

			case localTypes.Button_hall_down:
				if !elevio.IsHOrderActive(newBtnPress, CombinedHMatrix) {
					if len(localTypes.PeerList.Peers) == 0 {
						RxNewHallRequestChan <- newBtnPress
					} else {
						TxNewHallRequestChan <- newBtnPress
					}
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
			fmt.Printf("close door \n")
			elevio.SetDoorOpenLamp(false)
			newDir, newState := elevio.FindDirection2(MyElev, MyOrders)
			MyElev.Direction = newDir
			MyElev.State = newState
			elevio.SetMotorDirection(newDir)
			if newState == localTypes.Door_open {
				fmt.Printf("Run Elevator: dooropen LOOPING!\n")
				elevio.SetDoorOpenLamp(true)
				dooropentimer = time.NewTimer(localTypes.OPEN_DOOR_TIME_sek * time.Second)
			}
			localTypes.SendlocalElevInfo(MyElev, RxElevInfoChan, TxElevInfoChan)

		case AllElevs = <-RxP2PElevInfoChan:
			fmt.Printf("LE:innitALLELEVS\n")
			AllElevs = elevio.AddLocalToForeignInfo(MyElev, AllElevs)
			TxP2PElevInfoChan <- AllElevs
		}
	}

}
