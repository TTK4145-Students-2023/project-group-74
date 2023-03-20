package DLOCC

import (
	"config"
	"localTypes"
	"time"
)

const ORDER_WATCHDOG_POLL_RATE = config.ORDER_WATCHDOG_POLL_RATE_MS * time.Millisecond

type HRAElevState struct {
	State       string                      `json:"behaviour"`
	Floor       int                         `json:"floor"`
	Direction   string                      `json:"direction"`
	CabRequests [localTypes.NUM_FLOORS]bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [localTypes.NUM_FLOORS][2]bool `json:"hallRequests"`
	States       map[string]HRAElevState        `json:"states"`
}

type orderAssignerBehavior int

const (
	OABehaviorMaster orderAssignerBehavior = iota
	OABehaviorSlave
)

type OAInputs struct {
	localIDch         <-chan string
	ordersFromNetwork <-chan HRAInput
	ordersFromMaster  <-chan []byte
	ordersToSlave     chan<- []byte
	localOrder        chan<- [localTypes.NUM_FLOORS][2]bool
}
