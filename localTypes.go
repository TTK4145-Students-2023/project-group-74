package localTypes 

import 
  //project config 
  "strconv"

// ----- CONSTANTS ------ // 
// Create an init file with the following constants
// Time
// RX_BUFFER 

const (
  NUM_BUTTONS = 3
  NUM_FLOORS  = 4
  NUM_ORDERS  = NUM_FLOORS * NUM_BUTTONS
  
  OPEN_DOOR_TIME_sek            = 3
  TRAVEL_TIME_sek               = 3
  MAX_TIME_TO_FINISH_ORDER      = 3*(NUM_FLOORS-1)*(TRAVEL_TIME*OPEN_DOOR_TIME)
)

// ----- TYPE DEFINITIONS ------ // 



type BUTTON_TYPE int 
const(
  Button_Cab      BUTTON_TYPE = 0
  Button_hall_up              = 1
  Buton_hall_down            = 2
)

type BUTTON_INFO struct{
  Floor   int
  Button  BUTTON_TYPE
}

type HMATRIX [NUM_FLOORS][types.NUM_BUTTONS-1]bool
type ORDER map[string][types.NUM_FLOORS][types.NUM_BUTTONS-1]bool
type P2P_ELEV_INFO []LOCAL_ELEVATOR_INFO


type FOREIGN_ORDER_TYPE struct{
  Foregin_order BUTTON_INFO 
  Active        bool 
  Local         bool 
  New           bool 
}

type ELEVATOR_STATE int
const(
  idle      ELEVATOR_STATE = 0
  moving                   = 1
  door_open                = 2
)

type LOCAL_ELEVATOR_INFO struct{
  Floor       int 
  Direction   MOTOR_DIR 
  State       ELEVATOR_STATE
  CabCalls    [NUM_BUTTONS]bool
  ElevID      string   
}

type MOTOR_DIR int
const(
  DIR_down  MOTOR_DIR = -1
  DIR_stop            =  0
  DIR_up              =  1
)

//ack_foregin_elev?

// ----- FUNCTIONS (VALIDATION) ------ // 
func isValidFloor(floor int) bool{
  return floor>=0 && floor <= NUM_FLOORS
}

func isValidID(ID string) bool{
  id, err := strconv.Atoi(ID)
  if err != nil || id <0{
    return false}
  return true
}

func (state ELEVATOR_STATE) isValid() bool{
  return state == idle      ||
         state == moving    ||
         state == door_open 
}

func (button BUTTON_TYPE) isValid() bool{
  return button == Button_Cab        ||
         button == Button_hall_up    ||
         button == Button_hall_down  
}

func (btnInfo BUTTON_INFO) isValid() bool{
  return btnInfo.Button.isValid() && isValidFloor(btnInfo.Floor)
}

func (order ORDER_TYPE) isValid() bool{
  return BUTTON_INFO(order).isValid() 
}

func (dir MOTOR_DIR) isValid() bool{
  return dir == Dir_down  ||
         dir == Dir_up    ||
         dir == Dir_stop  
}

func (loc_elev LOCAL_ELEVATOR_INFO) isValid(){
  return isValidFloor(loc_elev.floor) &&
         loc_elev.Direction.isValid() &&
         loc_elev.State.isValid()     &&
  
}

// ----- FUNCTIONS (GET/SET) ------ // 
