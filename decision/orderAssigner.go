package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)

type HRAElevState struct {
	State        string              	`json:"behaviour"`
	Floor    	 int                 	`json:"floor"`
	Direction    string              	`json:"direction"`
	CabRequests  [types.NUM_FLOORS]bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [types.NUM_FLOORS][2]bool `json:"hallRequests"`
	States       map[string]HRAElevState   `json:"states"`
}

type orderAssignerBehavior int

const (
	OABehaviorMaster orderAssignerBehavior = iota
	OABehaviorSlave
)

type OAInputs struct {
	localIDch          <-chan 	string
	ordersFromNetwork  <-chan 	HRAInput
	ordersFromMaster   <-chan 	[]byte
	ordersToSlave        chan<- []byte
	localOrder           chan<- [types.NUM_FLOORS][2]bool
}

func orderAssigner(OA OAInputs) {
	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	mapOfElevators             := make(map[string]types.ExtendedElevator) //Define a map with information on all elevators
	mapOfElevators[elevator.ID] = elevator

	orderActivatedChn   := make(chan types.OrderType)
	orderDeactivatedChn := make(chan types.OrderType)
	orderTimedOutChn    := make(chan types.OrderType)


	localID := ""
	assignerBehavior := OABehaviorSlave
	//merge new hall-requests 
		//update buttons true/false
	//disconnected elevator 
	//foregin elevator 
	//local elevator -> prioritize cab 
	for {
		select {
		case localID = <-OA.localIDch:
		case assignerBehavior = <-orderAssignerBehaviorCh: // define this channel
		case givenOrders := <-OA.ordersFromNetwork:
			fmt.Printf("")
			switch assignerBehavior {
			case OABehaviorSlave: //This is info from the slaves, master vil send out 
			case OABehaviorMaster:
				jsonBytes, err := json.Marshal(givenOrders)
				if err != nil {
					fmt.Println("json.Marshal error: ", err)
					return
				}

				ret, err := exec.Command("../hall_request_assigner/"+ hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
				if err != nil {
					fmt.Println("exec.Command error: ", err)
					fmt.Println(string(ret))
					return
				}

				output := map[string][types.NUM_FLOORS][2]bool{}
				err = json.Unmarshal(ret, &output)
				if err != nil {
					fmt.Println("json.Unmarshal error: ", err)
					return
				}

				if localHallOrders, ok := output[localID]; ok {
					OA.localOrder <- localHallOrders
				}

				OA.ordersToSlave <- ret
			}
		case givenOrders := <-OA.ordersFromMaster:
			switch assignerBehavior {
			case OABehaviorMaster: //Do nothing since this is from master
			case OABehaviorSlave:
				output := map[string][types.NUM_FLOORS][2]bool{}
				err := json.Unmarshal(givenOrders, &output)
				if err != nil {
					fmt.Println("json.Unmarshal error: ", err)
					return
				}

				if localHallOrders, ok := output[localID]; ok {
					OA.localOrder <- localHallOrders
				}
			}
		}
	}
}
