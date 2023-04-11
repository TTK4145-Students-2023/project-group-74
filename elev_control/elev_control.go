package elev_control

import (
	"fmt"
	"project-group-74/elev_control/elevio"
	"project-group-74/localTypes"
	//"time"
)

//Channels:
//Output : To DLOCC : ElevInfoChan,NewHallRequestChan,FinishedHOrderChan
// To P2P   : TxP2PElevInfoChan
//Inputs : From DLOCC    : NewOrdersChan
// From Hardware : NewBtnPressChan, NewFloorChan
// From P2P      : RxP2pElevInfoChan

func RunElevator(
	TxElevInfoChan chan<- localTypes.LOCAL_ELEVATOR_INFO,
	RxElevInfoChan chan<- localTypes.LOCAL_ELEVATOR_INFO,
	TxNewHallRequestChan chan<- localTypes.BUTTON_INFO,
	RxNewHallRequestChan chan<- localTypes.BUTTON_INFO,
	TxFinishedHallOrderChan chan<- localTypes.BUTTON_INFO,
	RxFinishedHallOrderChan chan<- localTypes.BUTTON_INFO,
	RxNewOrdersChan <-chan map[string][localTypes.NUM_FLOORS][localTypes.NUM_BUTTONS - 1]bool,
	TxP2PElevInfoChan chan<- localTypes.P2P_ELEV_INFO,
	RxP2PElevInfoChan <-chan localTypes.P2P_ELEV_INFO,
	NewFloorChan chan int,
	NewBtnPressChan <-chan localTypes.BUTTON_INFO) {

	MyElev :=
		localTypes.LOCAL_ELEVATOR_INFO{
			State:     localTypes.Idle,
			Floor:     elevio.GetFloor(),
			Direction: localTypes.DIR_stop,
			CabCalls:  [localTypes.NUM_FLOORS]bool{},
			ElevID:    localTypes.MyIP,
		}

	MyElevPtr := &MyElev
	elevio.LocalElevInitFloor(&MyElev)

	var MyOrders localTypes.HMATRIX        // FOR DRIVING combined with myCabCalls
	var CombinedHMatrix localTypes.HMATRIX // FOR LIGHTS and reboot if you become master
	ForeignElevs := make(localTypes.P2P_ELEV_INFO, 0)
	ForeignElevsPtr := &ForeignElevs
	//var timeOutTimer = time.Now()
	//	p2pTicker := time.NewTicker(localTypes.P2P_UPDATE_INTERVAL * time.Millisecond)
	elevio.UpdateOrderLights(MyElev, CombinedHMatrix)

	for {
		select {
		case newOrder := <-RxNewOrdersChan:
			fmt.Printf("  neworder\n")
			elevio.AddNewOrders(newOrder, &MyOrders, &CombinedHMatrix, MyElev)
			elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
			TxP2PElevInfoChan <- ForeignElevs

		case newFloor := <-NewFloorChan:
			elevio.SetFloorIndicator(newFloor)
			fmt.Printf("  arrived at new floor:    %v\n", newFloor)
			MyElev.Floor = newFloor
			if elevio.IsOrderAtFloor(MyElev, MyOrders) == true {
				fmt.Printf("  order at new floor:    %v\n", newFloor)

				finishedOrder := elevio.GetFinOrder(newFloor, MyElev.Direction)
				if finishedOrder.Button == localTypes.Button_Cab {
					elevio.RemoveOneOrderBtn(finishedOrder, MyElevPtr)
					elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
				} else {
					if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) == true {
						fmt.Printf("RunElevator: New Floor: finished hall order\n")
						RxFinishedHallOrderChan <- finishedOrder
					} else {
						TxFinishedHallOrderChan <- finishedOrder
					}
				}
				elevio.ArrivedAtOrder(MyElevPtr) //Opendoors, wait, wait for them to press cab etc
			}
			fmt.Printf("  order not at new floor:    %v\n", newFloor)
			fmt.Printf("  redeciding locally\n")
			newDir, newState := elevio.FindDirection(MyElev, MyOrders)
			MyElev.Direction = newDir
			MyElev.State = newState
			elevio.SetMotorDirection(newDir)
			if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) == true {
				RxElevInfoChan <- MyElev
			} else {
				TxElevInfoChan <- MyElev
			}

		case newBtnPress := <-NewBtnPressChan:
			fmt.Printf("  Newbtnpress  \n ")
			if newBtnPress.Button == localTypes.Button_Cab {
				elevio.AddOneNewOrderBtn(newBtnPress, MyElevPtr)
				elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
				fmt.Printf("  redeciding locally\n")
				newDir, newState := elevio.FindDirection(MyElev, MyOrders)
				MyElev.Direction = newDir
				MyElev.State = newState
				elevio.SetMotorDirection(newDir)
				if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) == true {
					RxElevInfoChan <- MyElev
				} else {
					TxElevInfoChan <- MyElev
				}
			} else {
				if !elevio.IsHOrderActive(newBtnPress, CombinedHMatrix) {
					if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) == true {
						fmt.Printf("Run Elevator: newBtnPress: new hall request!\n")
						RxNewHallRequestChan <- newBtnPress
					} else {
						fmt.Printf("Run Elevator: newBtnPress: Broadcast new hall request!\n")
						TxNewHallRequestChan <- newBtnPress
					}

				}
			}

		case NewForeignInfo := <-RxP2PElevInfoChan:
			fmt.Printf("  neewforeign \n")

			ForeignElevs = NewForeignInfo
			elevio.AddLocalToForeignInfo(MyElev, ForeignElevsPtr)

			fmt.Printf("  sending ForeignElev: \n")
			TxP2PElevInfoChan <- ForeignElevs

			// case timer := <-p2pTicker.C:
			// 	fmt.Printf("  timer %v\n", timer)

			// 	elevio.AddLocalToForeignInfo(MyElev, ForeignElevsPtr)
			// 	fmt.Printf("  sending ForeignElev: \n")
			// 	TxP2PElevInfoChan <- ForeignElevs
		}
	}

}
