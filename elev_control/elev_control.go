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
			if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) {
				RxElevInfoChan <- MyElev
			} else {
				TxElevInfoChan <- MyElev
			}
			initializing = false
			fmt.Printf("Initializing finished!\n")
		default:
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

			switch MyElev.State {
			case localTypes.Door_open:
			case localTypes.Moving:
			case localTypes.Idle:
				if elevio.IsOrderAtFloor(MyElev, MyOrders) {
					MyElev.CabCalls[MyElev.Floor] = false
					if MyOrders[MyElev.Floor][localTypes.Button_hall_up] {
						if len(localTypes.PeerList.Peers) == 0 {
							RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
						} else {
							TxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
						}
					} else if MyOrders[MyElev.Floor][localTypes.Button_hall_down] {
						if len(localTypes.PeerList.Peers) == 0 {
							RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_down}
						} else {
							RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_down}
						}
					}
					MyElev.State = localTypes.Door_open
					elevio.SetDoorOpenLamp(true)
					fmt.Printf("Dooropen from neworder\n")
					dooropentimer = time.NewTimer(localTypes.OPEN_DOOR_TIME_sek * time.Second)
					if len(localTypes.PeerList.Peers) == 0 {
						RxElevInfoChan <- MyElev
					} else {
						TxElevInfoChan <- MyElev
					}
				} else {
					newDir, newState := elevio.FindDirection(MyElev, MyOrders)
					MyElev.Direction, MyElev.State = newDir, newState
					elevio.SetMotorDirection(newDir)
					if len(localTypes.PeerList.Peers) == 0 {
						RxElevInfoChan <- MyElev
					} else {
						TxElevInfoChan <- MyElev
					}
				}
			}
		case newFloor := <-NewFloorChan:
			if MyElev.Floor != newFloor {
				MyElev.Floor = newFloor
				elevio.SetFloorIndicator(MyElev.Floor)
			}
			switch MyElev.Direction {
			case localTypes.DIR_stop:
			case localTypes.DIR_up:
				switch {
				case MyOrders[MyElev.Floor][localTypes.Button_hall_up]:
					elevio.SetMotorDirection(localTypes.DIR_stop)
					MyElev.Direction = localTypes.DIR_stop
					MyElev.State = localTypes.Door_open
					MyElev.CabCalls[MyElev.Floor] = false

					if len(localTypes.PeerList.Peers) == 0 {
			 			RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
			 		} else {
			 			RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
			 		}
				case MyOrders[MyElev.Floor][localTypes.Button_hall_down]:
					elevio.SetMotorDirection(localTypes.DIR_stop)
					MyElev.Direction = localTypes.DIR_stop
					MyElev.State = localTypes.Door_open
					MyElev.CabCalls[MyElev.Floor] = false
					if len(localTypes.PeerList.Peers) == 0 {
						RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
					} else {
						RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
					}
				case elevio.Requests_above(MyElev, MyOrders):
				}
			case localTypes.DIR_down:
				switch {
					case MyOrders[MyElev.Floor][localTypes.Button_hall_down]:
						elevio.SetMotorDirection(localTypes.DIR_stop)
						MyElev.Direction = localTypes.DIR_stop
						MyElev.State = localTypes.Door_open
						MyElev.CabCalls[MyElev.Floor] = false

						if len(localTypes.PeerList.Peers) == 0 {
							RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
						} else {
							RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
						}
					case MyOrders[MyElev.Floor][localTypes.Button_hall_up]:
						elevio.SetMotorDirection(localTypes.DIR_stop)
						MyElev.Direction = localTypes.DIR_stop
						MyElev.State = localTypes.Door_open
						MyElev.CabCalls[MyElev.Floor] = false

						if len(localTypes.PeerList.Peers) == 0 {
							RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
						} else {
							RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
						}
					case elevio.Requests_below(MyElev, MyOrders):
				}
			}

			// switch true {
			// case MyElev.Direction == localTypes.DIR_up && MyOrders[MyElev.Floor][localTypes.Button_hall_up]:
			// 	MyElev.CabCalls[MyElev.Floor] = false
			// 	MyElev.Direction = localTypes.DIR_stop
			// 	MyElev.State = localTypes.Door_open
			// 	elevio.SetMotorDirection(localTypes.DIR_stop)
			// 	if len(localTypes.PeerList.Peers) == 0 {
			// 		RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
			// 	} else {
			// 		RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
			// 	}

			// case MyElev.Direction == localTypes.DIR_up && elevio.Requests_above(MyElev, MyOrders):

			// case MyElev.Direction == localTypes.DIR_up && MyOrders[MyElev.Floor][localTypes.Button_hall_down]:
			// 	MyElev.CabCalls[MyElev.Floor] = false
			// 	MyElev.Direction = localTypes.DIR_stop
			// 	MyElev.State = localTypes.Door_open
			// 	elevio.SetMotorDirection(localTypes.DIR_stop)
			// 	if len(localTypes.PeerList.Peers) == 0 {
			// 		RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_down}
			// 	} else {
			// 		RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_down}
			// 	}

			// case MyElev.Direction == localTypes.DIR_down && MyOrders[MyElev.Floor][localTypes.Button_hall_down]:
			// 	MyElev.CabCalls[MyElev.Floor] = false
			// 	MyElev.Direction = localTypes.DIR_stop
			// 	MyElev.State = localTypes.Door_open
			// 	elevio.SetMotorDirection(localTypes.DIR_stop)
			// 	if len(localTypes.PeerList.Peers) == 0 {
			// 		RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_down}
			// 	} else {
			// 		RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_down}
			// 	}

			// case MyElev.Direction == localTypes.DIR_down && elevio.Requests_below(MyElev, MyOrders):

			// case MyElev.Direction == localTypes.DIR_down && MyOrders[MyElev.Floor][localTypes.Button_hall_up]:
			// 	MyElev.CabCalls[MyElev.Floor] = false
			// 	MyElev.Direction = localTypes.DIR_stop
			// 	MyElev.State = localTypes.Door_open
			// 	elevio.SetMotorDirection(localTypes.DIR_stop)
			// 	if len(localTypes.PeerList.Peers) == 0 {
			// 		RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
			// 	} else {
			// 		RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
			// 	}

			// case MyElev.CabCalls[MyElev.Floor]:
			// 	MyElev.CabCalls[MyElev.Floor] = false
			// 	MyElev.Direction = localTypes.DIR_stop
			// 	MyElev.State = localTypes.Door_open
			// 	elevio.SetMotorDirection(localTypes.DIR_stop)

			// default:
			// 	MyElev.Direction = localTypes.DIR_stop
			// 	MyElev.State = localTypes.Idle
			// 	elevio.SetMotorDirection(localTypes.DIR_stop)
			// }

			// elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
			// if len(localTypes.PeerList.Peers) == 0 {
			// 	RxElevInfoChan <- MyElev
			// } else {
			// 	TxElevInfoChan <- MyElev
			// }

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
						MyElev.CabCalls = elevio.RemoveOneOrderBtn(newBtnPress, MyElev)
						elevio.UpdateOrderLights(MyElev, CombinedHMatrix)

						dooropentimer = time.NewTimer(localTypes.OPEN_DOOR_TIME_sek * time.Second)
					}
					if len(localTypes.PeerList.Peers) == 0 {
						RxElevInfoChan <- MyElev
					} else {
						TxElevInfoChan <- MyElev
					}
				}
			case localTypes.Button_hall_up, localTypes.Button_hall_down:
				if !elevio.IsHOrderActive(newBtnPress, CombinedHMatrix) {
					if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) {
						fmt.Printf("Run Elevator: new hall request!\n")

						RxNewHallRequestChan <- newBtnPress
					} else {
						RxNewHallRequestChan <- newBtnPress
					}
				}
			}

		case <-ObstructionChan:
			if MyElev.State == localTypes.Door_open {
				dooropentimer.Reset(localTypes.OPEN_DOOR_TIME_sek * time.Second)
			}

		case <-dooropentimer.C:

			if MyElev.State == localTypes.Door_open {
				elevio.SetDoorOpenLamp(false)
				newDir, newState := elevio.FindDirection(MyElev, MyOrders)
				MyElev.Direction = newDir
				MyElev.State = newState
				elevio.SetMotorDirection(newDir)
				if newState == localTypes.Door_open {
					fmt.Printf("Run Elevator: dooropen LOOPING!\n")
					fmt.Print("LOOPING ORDER %+b", MyElev.CabCalls[MyElev.Floor])

					elevio.SetDoorOpenLamp(true)
					dooropentimer = time.NewTimer(localTypes.OPEN_DOOR_TIME_sek * time.Second)
				}
				if len(localTypes.PeerList.Peers) == 0 {
					RxElevInfoChan <- MyElev
				} else {
					TxElevInfoChan <- MyElev
				}
			}

		case AllElevs = <-RxP2PElevInfoChan:
			fmt.Printf("LE:innitALLELEVS\n")
			AllElevs = elevio.AddLocalToForeignInfo(MyElev, AllElevs)
			TxP2PElevInfoChan <- AllElevs
		}
	}

}
