package DLOCC

import (
	"project-group-74/localTypes"
	"time"
)

const ORDER_WATCHDOG_POLL_RATE = 50 * time.Millisecond

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

var motorDirStrings = map[localTypes.MOTOR_DIR]string{
	localTypes.DIR_down: "down",
	localTypes.DIR_stop: "stop",
	localTypes.DIR_up:   "up",
}

var elevStateStrings = map[localTypes.ELEVATOR_STATE]string{
	localTypes.Idle:      "idle",
	localTypes.Moving:    "moving",
	localTypes.Door_open: "doorOpen",
}
