package elev_control

import (
	"fmt"
	"project-group-74/elev_control/elevio"
	"project-group-74/localTypes"
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

	//MyElevPtr := &MyElev Remove pointers
	MyElev = elevio.LocalElevInitFloor(MyElev)

	var MyOrders localTypes.HMATRIX        // FOR DRIVING combined with myCabCalls
	var CombinedHMatrix localTypes.HMATRIX // FOR LIGHTS and reboot if you become master
	ForeignElevs := make(localTypes.P2P_ELEV_INFO, 0)
	//ForeignElevsPtr := &ForeignElevs Remove pointers
	//var timeOutTimer = time.Now()
	//	p2pTicker := time.NewTicker(localTypes.P2P_UPDATE_INTERVAL * time.Millisecond)

	elevio.UpdateOrderLights(MyElev, CombinedHMatrix)

	for {
		select {
		case newOrder := <-RxNewOrdersChan:

			MyOrders = elevio.AddNewOrdersToLocal(newOrder, MyOrders, MyElev)
			CombinedHMatrix = elevio.AddNewOrdersToHMatrix(newOrder)

			elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
			TxP2PElevInfoChan <- ForeignElevs
			newDir, newState := elevio.FindDirection(MyElev, MyOrders)
			MyElev.Direction = newDir
			MyElev.State = newState
			elevio.SetMotorDirection(newDir)
			if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) {
				RxElevInfoChan <- MyElev
			} else {
				TxElevInfoChan <- MyElev
			}

		case newFloor := <-NewFloorChan:
			elevio.SetFloorIndicator(newFloor)
			MyElev.Floor = newFloor

			if elevio.IsOrderAtFloor(MyElev, MyOrders) {
				finishedOrder := elevio.GetFinOrder(newFloor, MyElev.Direction)
				if finishedOrder.Button == localTypes.Button_Cab {
					MyElev.CabCalls = elevio.RemoveOneOrderBtn(finishedOrder, MyElev)
					elevio.UpdateOrderLights(MyElev, CombinedHMatrix)
					fmt.Printf("LE:finished cab order\n")

				} else {

					if localTypes.IsMaster(MyElev.ElevID, localTypes.PeerList.Peers) {
						fmt.Printf("LE:finished hall order\n")
						RxFinishedHallOrderChan <- finishedOrder
					} else {
						TxFinishedHallOrderChan <- finishedOrder
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

		case NewForeignInfo := <-RxP2PElevInfoChan:
			ForeignElevs = NewForeignInfo

			ForeignElevs = elevio.AddLocalToForeignInfo(MyElev, ForeignElevs)
			TxP2PElevInfoChan <- ForeignElevs

		default:

			// case timer := <-p2pTicker.C:
			// 	fmt.Printf("  timer %v\n", timer)

			// 	elevio.AddLocalToForeignInfo(MyElev, ForeignElevsPtr)
			// 	fmt.Printf("  sending ForeignElev: \n")
			// 	TxP2PElevInfoChan <- ForeignElevs
		}
	}

}
