package utilities

//Imports
import (
	"time"
)
//Constants
const(
	DEFAULT_IP			= "127.0.0.1"
)
const(
	DISCONNECTED		= 0
	CONNECTED 			= 1
)
const(
	ELEVATORS			= 2
)
const(
	INVALID 			=-1
	TRUE    			= 1
	FALSE   			= 0
)
const(
	ON      			= 1
	OFF     			= 0
)
const (
	STOP 				= 0
	UP   				= 2
	DOWN 				= 1
)
const (
	FLOORS  			= 4
	BUTTONS 			= 3 
)
const (
	FLOOR_FIRST  		= 0
	FLOOR_SECOND 		= 1
	FLOOR_THIRD  		= 2
	FLOOR_LAST 			= 3
)
const (
	BUTTON_INVALID  	=-1
	BUTTON_UP     		= 0
	BUTTON_DOWN   		= 1
	BUTTON_INSIDE 		= 2
	BUTTON_OUTSIDE		= 3
	BUTTON_STOP  		= 4
	BUTTON_OBSTRUCTION 	= 5
)

const (
	STATE_INVALID	 	=-1
	STATE_STARTUP  		= 0
	STATE_IDLE	  		= 1
	STATE_RUNNING  		= 2
	STATE_EMERGENCY		= 3
	STATE_DOOR_OPEN		= 4
	STATE_DOOR_CLOSED	= 5
)

const (
	MESSAGE_INVALID	 	=-1
	MESSAGE_HEARTBEAT	= 0
	MESSAGE_STATUS 		= 1
	MESSAGE_ORDER		= 2
	MESSAGE_FULFILLED	= 3
	MESSAGE_ORDERS		= 4
	MESSAGE_REQUEST		= 5
	MESSAGE_REPRIORITIZE= 6
)

//Types 
type ID string

type Message struct {
	Category 	int
	Heartbeat 	Heartbeat
	Status 		Status
	Order		Order
	Orders 		Orders
}

//Structs
type Heartbeat struct {
	Counter int
}
type Status struct {
	Elevator 	ID
	State 		int
	Floor 		int
	Direction 	int
	Priority	int
	Time		time.Time
}
type Orders struct{
	Elevator 	ID
	List [] 	Order 
}
type Priority struct{
	Elevator 	ID
	Count		int
}
type Order struct {
	Elevator 	ID 
	Category	int 
	Direction	int
	Floor 		int  
	Button      int
	Time		time.Time
}
type Floor struct {
	Current int
	Status int
}
type Packet struct {
	Address 	ID 
	Data    	[]byte
}
