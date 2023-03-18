package order_assigner

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)

const ORDER_WATCHDOG_POLL_RATE = config.ORDER_WATCHDOG_POLL_RATE_MS * time.Millisecond

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
