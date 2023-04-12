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
	NewFloorChan chan int,
	NewBtnPressChan <-chan localTypes.BUTTON_INFO) {

	MyElev :=
		localTypes.LOCAL_ELEVATOR_INFO{
			State:     localTypes.Idle,
			Floor:     -1,
			Direction: localTypes.DIR_stop,
			CabCalls:  [localTypes.NUM_FLOORS]bool{},
			ElevID:    myIP,
		}

	//MyElevPtr := &MyElev Remove pointers
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
	fmt.Printf("LE:innit\n")

	elevio.UpdateOrderLights(MyElev, CombinedHMatrix)

	for {
		select {
		case newOrder := <-RxNewOrdersChan:
			MyOrders = newOrder[myIP]

			for _, orders := range newOrder {
				for floor, buttons := range orders {
					for button, val := range buttons {
						CombinedHMatrix[floor][button] = CombinedHMatrix[floor][button] || val
					}
				}
			}

			elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
			TxP2PElevInfoChan <- AllElevs
			newDir, newState := elevio.FindDirection(MyElev, MyOrders)
			MyElev.Direction, MyElev.State = newDir, newState
			elevio.SetMotorDirection(newDir)

			if newState == localTypes.Door_open {
				MyElev.CabCalls[MyElev.Floor] = false
				elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
				if MyOrders[MyElev.Floor][localTypes.Button_hall_up] {
					if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) {
						RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
					} else {
						TxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
					}
				} else if MyOrders[MyElev.Floor][localTypes.Button_hall_down] {
					if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) {
						RxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_down}
					} else {
						TxFinishedHallOrderChan <- localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_down}
					}
				}
				MyElev = elevio.ArrivedAtOrder(MyElev) //Opendoors, wait
			}
			if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) {
				RxElevInfoChan <- MyElev
			} else {
				TxElevInfoChan <- MyElev
			}

		case newFloor := <-NewFloorChan:
			fmt.Printf("LE:NewFLoorProc\n")
			elevio.SetFloorIndicator(newFloor)
			MyElev.Floor = newFloor

			if elevio.IsOrderAtFloor(MyElev, MyOrders) {
				MyElev.CabCalls[MyElev.Floor] = false
				elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
				var completedorder localTypes.BUTTON_INFO
				var changed bool
				switch MyElev.Direction {
				case localTypes.DIR_up:
					if elevio.Requests_above(MyElev, MyOrders) {
						if MyOrders[MyElev.Floor][localTypes.Button_hall_up] {
							completedorder = localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
							changed = true
						}
					} else {
						if MyOrders[MyElev.Floor][localTypes.Button_hall_down] {
							completedorder = localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_down}
							changed = true
						}
					}
				case localTypes.DIR_down:
					if elevio.Requests_below(MyElev, MyOrders) {
						if MyOrders[MyElev.Floor][localTypes.Button_hall_down] {
							completedorder = localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_down}
							changed = true
						}
					} else {
						if MyOrders[MyElev.Floor][localTypes.Button_hall_up] {
							completedorder = localTypes.BUTTON_INFO{Floor: MyElev.Floor, Button: localTypes.Button_hall_up}
							changed = true
						}
					}
				}
				if changed {
					if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) {
						RxFinishedHallOrderChan <- completedorder
					} else {
						TxFinishedHallOrderChan <- completedorder
					}
				}
				MyElev = elevio.ArrivedAtOrder(MyElev) //Opendoors, wait, wait for them to press cab etc
			}
			fmt.Printf("  redeciding locally-newfloor\n")
			newDir, newState := elevio.FindDirection(MyElev, MyOrders)
			MyElev.Direction = newDir
			MyElev.State = newState
			elevio.SetMotorDirection(newDir)
			if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) {
				RxElevInfoChan <- MyElev
			} else {
				TxElevInfoChan <- MyElev
			}

		case newBtnPress := <-NewBtnPressChan:
			fmt.Printf("LE:NewBTNProc\n")

			if newBtnPress.Button == localTypes.Button_Cab {
				fmt.Printf("Run Elevator: new cab request!\n")
				MyElev.CabCalls = elevio.AddOneNewOrderBtn(newBtnPress, MyElev)
				elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
				newDir, newState := elevio.FindDirection(MyElev, MyOrders)
				MyElev.Direction = newDir
				MyElev.State = newState
				elevio.SetMotorDirection(newDir)
				if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) {
					RxElevInfoChan <- MyElev
				} else {
					TxElevInfoChan <- MyElev
				}
			} else {
				if !elevio.IsHOrderActive(newBtnPress, CombinedHMatrix) {

					if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) {
						fmt.Printf("Run Elevator: new hall request!\n")

						RxNewHallRequestChan <- newBtnPress
					} else {
						TxNewHallRequestChan <- newBtnPress
					}

				}
			}

		case AllElevs = <-RxP2PElevInfoChan:
			fmt.Printf("LE:Newp2pelev\n")


			AllElevs = elevio.AddLocalToForeignInfo(MyElev, AllElevs)
			TxP2PElevInfoChan <- AllElevs
		}
	}

}
