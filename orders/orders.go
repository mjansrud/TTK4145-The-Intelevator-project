package orders

import (
	"../driver"
	"../network"
	"../states"
	"../utilities"
	"fmt"
	"math"
	"reflect"
	"sync"
	"time"
)

//Variables
var filename string = "Orders -"
var mutex sync.Mutex
var inserting bool
var orders []utilities.Order
var orders_offline []utilities.Order
var check_counter = 0

func Handle(channel_poll_floor chan utilities.Floor, channel_poll_order chan utilities.Order, channel_write chan utilities.Packet) {

	for {
		select {
		//Called when we have a new order
		case current_order := <-channel_poll_order:

			//Prevent accessing the same memory at the same time
			if !inserting {

				inserting = true
				PrioritizeOrder(&current_order)

				//We we not want a duplicate order in the system
				if !CheckOrderExists(current_order) {

					//Network messages
					var message_send utilities.Message
					var packet_send utilities.Packet

					//Create and send message
					message_send.Category = utilities.MESSAGE_ORDER
					message_send.Order = current_order
					packet_send.Data = network.EncodeMessage(message_send)
					channel_write <- packet_send

					//Insert order into local array
					InsertOrder(current_order)

				}

				inserting = false

			}
		//Called when we reach a new floor
		case current_floor := <-channel_poll_floor:

			//Update floor
			states.SetFloor(current_floor.Current)
			CheckOrdersCompleted(channel_write)
			RunActiveOrders()

			//Check if we are at bottom
			if states.GetFloor() == utilities.FLOOR_FIRST {
				driver.DirectionUp()
			}

			//Check if we are at top
			if states.GetFloor() == utilities.FLOOR_LAST {
				driver.DirectionDown()
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}
func PrioritizeState(elevator_state int) int {
	if elevator_state == utilities.STATE_IDLE {
		return 1
	}
	return 0
}

func PrioritizeFloor(current_floor int, order_floor int) int {
	return utilities.FLOORS - int(math.Abs(float64(current_floor-order_floor)))
}

func PrioritizeDirection(elevator_state int, elevator_direction int, order_direction int) int {
	if elevator_state == utilities.STATE_RUNNING {
		if elevator_direction == order_direction {
			return 1
		}
	}
	return 0
}

func PrioritizeOrders() {

	for index := range orders {
		PrioritizeOrder(&orders[index])
	}

	for index := range orders_offline {
		PrioritizeOrder(&orders_offline[index])
	}

	PrintOrders()
	fmt.Println(filename, "All orders reprioritized")
}

func PrioritizeOrder(order *utilities.Order) {

	copy := GetOrder(*order)

	if copy.Category == utilities.BUTTON_INSIDE {
		if copy.Elevator == states.GetConnectedIp() || copy.Elevator == utilities.DEFAULT_IP {

			//Check network connection
			if states.IsConnected() {
				mutex.Lock()
				order.Elevator = states.GetConnectedIp()
				mutex.Unlock()
			}

			//If disconnected
			if !states.IsConnected() {
				mutex.Lock()
				order.Elevator = utilities.DEFAULT_IP
				mutex.Unlock()
			}
		}
	}

	//Only prioritize orders pushed outside of elevators or local orders pushed inside
	if copy.Category == utilities.BUTTON_OUTSIDE {

		var priority utilities.Priority
		var priorities []utilities.Priority
		var elevators []utilities.Status = states.GetExternalElevators()

		//Local elevator
		priority.Elevator = network.GetMachineID()

		//Prioritize
		priority.Count += PrioritizeState(states.GetState())
		priority.Count += PrioritizeFloor(states.GetFloor(), copy.Floor)
		priority.Count += PrioritizeDirection(states.GetState(), states.GetDirection(), copy.Direction)

		//Add priority
		priorities = append(priorities, priority)

		//External elevators
		if len(elevators) > 0 {

			for index := range elevators {

				var priority utilities.Priority

				//Fetch elevator
				priority.Elevator = elevators[index].Elevator

				//Calculate priority
				priority.Count += PrioritizeState(elevators[index].State)
				priority.Count += PrioritizeFloor(elevators[index].Floor, copy.Floor)
				priority.Count += PrioritizeDirection(elevators[index].State, elevators[index].Direction, copy.Direction)

				//Add priority
				priorities = append(priorities, priority)

			}
		}

		//All priorities
		if len(priorities) > 0 {

			for index := range priorities {

				//If we have a bigger score
				if priorities[index].Count > priority.Count {

					//Change elevator
					priority.Elevator = priorities[index].Elevator
					priority.Count = priorities[index].Count

				}

				//If we have the same score - compare IP's. Biggest IP gets order.
				if priorities[index].Count == priority.Count {

					//Compare
					if priorities[index].Elevator > priority.Elevator {

						//Change elevator
						priority.Elevator = priorities[index].Elevator
						priority.Count = priorities[index].Count

					}
				}

			}
		}

		//Set target elevator to the elevator most appropriate
		mutex.Lock()
		order.Elevator = priority.Elevator
		mutex.Unlock()

	}

}

func InsertOrder(order utilities.Order) {

	//Check if it is a order pushed inside or the correct direction outside
	if (order.Elevator == network.GetMachineID() && order.Category == utilities.BUTTON_INSIDE) || (order.Category == utilities.BUTTON_OUTSIDE) {

		driver.SetButtonLamp(order.Button, order.Floor, utilities.ON)

	}

	//Add order
	mutex.Lock()
	orders = append(orders, order)
	mutex.Unlock()

	//Successfully added
	PrintOrder(order)

}

func InsertOfflineOrder(order utilities.Order) {

	local_offline := GetOfflineOrders()
	found := false

	if len(local_offline) > 0 {
		for index := range local_offline {

			//Check if all fields are alike
			if reflect.DeepEqual(local_offline[index], order) {

				found = true

			}
		}

	}

	//Add order
	if !found && order.Category == utilities.BUTTON_INSIDE {

		mutex.Lock()
		orders_offline = append(orders_offline, order)
		mutex.Unlock()

		//Successfully added
		PrintOrder(order)

	}

}

func CountRelevantOrders() int {

	counter := 0
	local_orders := GetOrders()

	if len(local_orders) > 0 {
		for index := range local_orders {

			//Check if the order has something to do with local machine
			if local_orders[index].Elevator == network.GetMachineID() {
				//Count
				counter++

			}
		}
	}
	return counter
}

func RunActiveOrders() {

	//Check if we have intiated an order
	if states.GetState() != utilities.STATE_DOOR_OPEN {

		if CountRelevantOrders() > 0 {
			//Initiate variables
			orders_over := false
			orders_under := false
			local_orders := GetOrders()

			//Check if we have orders above or under current floor
			for index := range local_orders {

				//Check if it is a order pushed inside or the correct direction outside
				if local_orders[index].Elevator == network.GetMachineID() {

					if local_orders[index].Floor > states.GetFloor() {
						orders_over = true
					}

					if local_orders[index].Floor < states.GetFloor() {
						orders_under = true
					}
				}

			}

			//Run elevator in the correct direction
			if orders_over && !(orders_under && states.GetDirection() == utilities.DOWN) {
				driver.RunUp()
			} else if orders_under && !(orders_over && states.GetDirection() == utilities.UP) {
				driver.RunDown()
			}

			//If we dont have orders over or under, change direction
			if !orders_over && !orders_under {
				driver.DirectionSwitch()
			}

		} else {

			//We have no new relevant orders.
			driver.Stop()
			states.SetState(utilities.STATE_IDLE)

		}

	}

}

func CheckOrderExists(order utilities.Order) bool {

	local_orders := GetOrders()

	if len(local_orders) > 0 {
		for index := range local_orders {
			//Check if all fields are alike
			if reflect.DeepEqual(local_orders[index], order) {
				return true
			}
		}
	}

	return false

}

func CheckOrdersCompleted(channel_write chan utilities.Packet) {

	//Check if we have intiated an order
	if states.GetState() != utilities.STATE_DOOR_OPEN {

		if CountRelevantOrders() > 0 {

			local_orders := GetOrders()

			for index := range local_orders {

				//Check if we are on the same floor as the order
				if local_orders[index].Floor == states.GetFloor() {

					//Check if it is a order pushed inside or the correct direction outside
					if (local_orders[index].Elevator == network.GetMachineID()) && (local_orders[index].Category == utilities.BUTTON_INSIDE) || (local_orders[index].Category == utilities.BUTTON_OUTSIDE && local_orders[index].Direction == states.GetDirection()) {

						InitCompleteOrder(channel_write, local_orders[index])

						//Prevent panic
						break

					}
				}
			}
		}

	}
}

func InitCompleteOrder(channel_write chan utilities.Packet, order utilities.Order) {

	local_orders := GetOrders()

	for index := range local_orders {

		//Check if all fields are alike
		if reflect.DeepEqual(local_orders[index], order) {

			//Open door for the specific elevator
			if order.Elevator == network.GetMachineID() {

				//Order-being processed sequence
				driver.Stop()
				states.SetState(utilities.STATE_DOOR_OPEN)
				driver.SetDoorLamp(utilities.ON)
				timer := time.NewTimer(time.Second)

				//Finish order when timer is done
				go CompleteOrder(channel_write, order, timer)

			} else {

				driver.SetButtonLamp(order.Button, order.Floor, utilities.OFF)
				fmt.Println(filename, "Order completed")

			}

			//Prevent reconnecting elevator double order handling
			if !states.IsConnected() {
				InsertOfflineOrder(order)
			}

			//Remove order
			mutex.Lock()
			orders = append(orders[:index], orders[index+1:]...)
			mutex.Unlock()

			//Prevent panic
			break

		}

	}
}

func CompleteOrder(channel_write chan utilities.Packet, order utilities.Order, timer *time.Timer) {

	//When timer is finished
	<-timer.C

	//Order-complete sequence
	driver.SetButtonLamp(order.Button, order.Floor, utilities.OFF)
	driver.SetDoorLamp(utilities.OFF)
	states.SetState(utilities.STATE_DOOR_CLOSED)

	//Network messages
	var message_send utilities.Message
	var packet_send utilities.Packet

	//Create and send message
	message_send.Category = utilities.MESSAGE_FULFILLED
	message_send.Order = order
	packet_send.Data = network.EncodeMessage(message_send)
	channel_write <- packet_send

	fmt.Println(filename, "Order completed")

}

//Requesting orders from other computers on the local network
func RequestOrders(channel_write chan utilities.Packet) {

	fmt.Println(filename, "Requesting orders")

	//Network messages
	var message_send utilities.Message
	var packet_send utilities.Packet

	//Create and send message
	message_send.Category = utilities.MESSAGE_REQUEST
	packet_send.Data = network.EncodeMessage(message_send)
	channel_write <- packet_send

}

//Getters and setters
func GetOrder(order utilities.Order) utilities.Order {

	mutex.Lock()
	defer mutex.Unlock()

	//Make a full data copy
	o := utilities.Order{}
	o = order

	return o
}

func GetOrders() []utilities.Order {

	mutex.Lock()
	defer mutex.Unlock()

	//Create a copy - preventing data race
	o := make([]utilities.Order, len(orders), len(orders))

	//Need to manually copy all variables - Library "copy" function will not work
	for id, elem := range orders {
		o[id] = elem
	}

	return o
}

func GetOfflineOrders() []utilities.Order {

	mutex.Lock()
	defer mutex.Unlock()

	//Create a copy - preventing data race
	o := make([]utilities.Order, len(orders_offline), len(orders_offline))

	//Need to manually copy all variables - Library "copy" function will not work
	for id, elem := range orders_offline {
		o[id] = elem
	}

	return o
}

func RemoveOfflineHistory() {

	mutex.Lock()
	orders_offline = orders_offline[:0]
	mutex.Unlock()

}

func SetOrders(list []utilities.Order, channel_write chan utilities.Packet) {

	for index := range list {

		found := false
		for offline_index := range orders_offline {

			//Check if all fields are alike
			if reflect.DeepEqual(list[index], orders_offline[offline_index]) {

				found = true

			}
		}

		if !found {
			InsertOrder(list[index])
		}
	}
}

//Printing functionality
func PrintOrders() {

	local_orders := GetOrders()

	for index := range local_orders {
		PrintOrder(local_orders[index])
	}

}
func PrintOrder(order utilities.Order) {

	if order.Category == utilities.BUTTON_INSIDE {

		fmt.Println(filename, "Order inside, ip: ", order.Elevator, ",floor: ", order.Floor)

	}

	if order.Category == utilities.BUTTON_OUTSIDE {

		fmt.Print(filename, " Order outside, ip: ", order.Elevator, ",floor: ", order.Floor, ",direction: ")

		if order.Direction == utilities.BUTTON_UP {
			fmt.Print(" up")
		} else {
			fmt.Print(" down")
		}

		fmt.Println()

	}

}

func PrintPriority(local_priority utilities.Priority) {

	fmt.Println(filename, " Elevator: ", local_priority.Elevator, ", count: ", local_priority.Count)

}

func PrintPriorities(local_priorities []utilities.Priority) {

	for index := range local_priorities {
		PrintPriority(local_priorities[index])
	}

}
