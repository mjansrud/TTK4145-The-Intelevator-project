package events

import (
	"../driver"
	"../network"
	"../orders"
	"../states"
	"../utilities"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

var filename string = "Events -"
var heartbeat = time.Now()

func CheckOrderRequested() (int, int) {
	for FLOOR := 0; FLOOR < utilities.FLOORS; FLOOR++ {
		for BUTTON := 0; BUTTON < utilities.BUTTONS; BUTTON++ {
			if driver.GetButtonSignal(BUTTON, FLOOR) == utilities.TRUE {
				return BUTTON, FLOOR
			}
		}
	}
	return utilities.INVALID, utilities.INVALID
}

func PollFloor(ch_floor_poll chan utilities.Floor) {

	for {

		polled_floor := driver.GetFloorSignal()
		if polled_floor != utilities.INVALID {

			var new_floor utilities.Floor
			new_floor.Current = polled_floor
			driver.SetFloorIndicator(polled_floor)
			ch_floor_poll <- new_floor

		}

		time.Sleep(10 * time.Millisecond)
	}
}

func PollOrder(ch_order_poll chan utilities.Order) {

	for {
		// Get registered button and floor requested (see header.go)
		polled_order_button, polled_order_floor := CheckOrderRequested()

		// Only go through if not invalid button
		if polled_order_button != utilities.INVALID && polled_order_floor != utilities.INVALID {

			var new_order utilities.Order
			new_order.Elevator = network.GetMachineID()

			if polled_order_button == utilities.BUTTON_INSIDE {
				new_order.Category = utilities.BUTTON_INSIDE
			}
			if polled_order_button == utilities.BUTTON_UP {
				new_order.Category = utilities.BUTTON_OUTSIDE
				new_order.Direction = utilities.UP
			}
			if polled_order_button == utilities.BUTTON_DOWN {
				new_order.Category = utilities.BUTTON_OUTSIDE
				new_order.Direction = utilities.DOWN
			}

			//Register the new order
			new_order.Floor = polled_order_floor
			new_order.Button = polled_order_button
			new_order.Time = time.Now()
			ch_order_poll <- new_order
		}

		time.Sleep(10 * time.Millisecond)

	}
}

/*
	1) Function checks if we have a local or external IP which desides if we are connected to the network
	2) The function loops through the connected elevators and removed them if we havent received a heartbeat within reasonable time
	3) We loop through all orders and check if they have been executed within x seconds. If not, we assume the hardware/motor is broken.
*/
func StatusChecker(channel_write chan utilities.Packet) {

	//Network messages
	var message_send utilities.Message
	var packet_send utilities.Packet

	//Run loop forever
	for {

		//Computer is disconnected. Handle orders locally.
		if states.IsConnected() && network.GetMachineID() == utilities.DEFAULT_IP {

			states.SetConnection(utilities.DISCONNECTED)
			time.Sleep(1 * time.Second)
			orders.PrioritizeOrders()

		}

		//Computer is connected. Request orders from external elevators.
		if !states.IsConnected() && network.GetMachineID() != utilities.DEFAULT_IP {

			states.SetConnection(utilities.CONNECTED)
			states.SetConnectedIp(network.GetMachineID())
			time.Sleep(1 * time.Second)
			orders.PrioritizeOrders()
			orders.RequestOrders(channel_write)

		}

		local_elevators := states.GetExternalElevators()
		for elevator_index := range local_elevators {

			//Fetch specific elevator for easier syntax
			elevator := local_elevators[elevator_index]

			//Calculate time
			elapsed := time.Since(elevator.Time)
			elapsed = (elapsed + time.Second/2) / time.Second

			//Check for timeout
			if elapsed > 3 {
				states.RemoveExternalElevator(elevator)
				orders.PrioritizeOrders()
			}
		}

		if(states.GetState() != utilities.STATE_EMERGENCY){

			local_orders := orders.GetOrders()
			for orders_index := range local_orders {

				//Fetch specific order for easier syntax
				order := local_orders[orders_index]

				//Calculate time
				elapsed := time.Since(order.Time)
				elapsed = (elapsed + time.Second/2) / time.Second

				//Check for timeout 
				if order.Elevator == network.GetMachineID() && elapsed > 20 {
					
					states.SetState(utilities.STATE_EMERGENCY)
					orders.PrioritizeOrders()

					//Let the external pc's wait to receive a status message
					time.Sleep(1 * time.Second)

					//Create and send message
					message_send.Category = utilities.MESSAGE_REPRIORITIZE
					packet_send.Data = network.EncodeMessage(message_send)
					channel_write <- packet_send

					fmt.Println(filename, " Reprioritize request sent")
				}
			}

		}

		time.Sleep(1 * time.Second)

	}
}

/*
	Broadcasts a status message to the other elevators each 0.5 seconds.
*/
func NetworkLoop(channel_write chan utilities.Packet) {

	//Network messages
	var message_send utilities.Message
	var packet_send utilities.Packet

	//Run loop forever
	for {

		//Check if we are master
		if states.IsMaster() {

			//Create status message
			message_send.Category = utilities.MESSAGE_STATUS
			message_send.Status.Elevator = network.GetMachineID()
			message_send.Status.State = states.GetState()
			message_send.Status.Floor = states.GetFloor()
			message_send.Status.Direction = states.GetDirection()
			message_send.Status.Time = time.Now()

			//Encode and send
			packet_send.Data = network.EncodeMessage(message_send)
			channel_write <- packet_send

		}

		time.Sleep(500 * time.Millisecond)

	}

}

/*
	Function listens on the network and receives messages from the other elevators.
*/
func NetworkListener(channel_listen chan utilities.Packet, channel_write chan utilities.Packet) {

	//Network messages
	var message_receive utilities.Message
	var packet_receive utilities.Packet

	for {

		//Check if we are master
		if states.IsMaster() {

			select {

			case packet_receive = <-channel_listen:

				//Get message and decode
				message_receive = network.DecodeMessage(packet_receive.Data)

				//Do not treat own sent messages
				if packet_receive.Address != network.GetMachineID() {

					//Do something depending on type
					switch message_receive.Category {

					case utilities.MESSAGE_STATUS:

						states.UpdateExternalElevator(message_receive.Status)
						break

					case utilities.MESSAGE_ORDER:

						//Order received from another elevator
						fmt.Println(filename, "Order received")

						//Check if this is not the sending machine
						if !orders.CheckOrderExists(message_receive.Order) {
							orders.InsertOrder(message_receive.Order) 
						}
						break

					case utilities.MESSAGE_FULFILLED:

						//Order fulfilled by another elvator
						fmt.Println(filename, "Order fulfilled")
						orders.InitCompleteOrder(channel_write, message_receive.Order)
						break

					case utilities.MESSAGE_ORDERS:

						fmt.Println(filename, "Order list received")

						//Check if this is the intended machine
						if message_receive.Orders.Elevator == network.GetMachineID() {
							orders.SetOrders(message_receive.Orders.List, channel_write)
						}

						//Prevent reconnecting elevator double order handling
						orders.RemoveOfflineHistory()
						break

					case utilities.MESSAGE_REQUEST:

						//Requesting order list
						fmt.Println(filename, "Elevator ", packet_receive.Address, " requesting orders")

						//Network messages
						var message_send utilities.Message
						var packet_send utilities.Packet

						//Create and send message
						message_send.Category = utilities.MESSAGE_ORDERS
						message_send.Orders.Elevator = packet_receive.Address
						message_send.Orders.List = orders.GetOrders()
						packet_send.Data = network.EncodeMessage(message_send)
						channel_write <- packet_send

						fmt.Println(filename, " Orders sent")
						break

					case utilities.MESSAGE_REPRIORITIZE:

						//Order fulfilled by another elvator
						fmt.Println(filename, " Reprioritize")
						orders.PrioritizeOrders()
						break
					}
					
				}

			}

		}

		time.Sleep(10 * time.Millisecond)

	}

}

/*
	Function checks if we have received a heartbeat from the master process locally.
	If it is longer than 3 seconds since the last heartbeat, spawn as a master.
*/
func HeartbeatChecker(channel_backup_quit, channel_backup_init_master chan bool) {

	//Loop
	for {

		//Check if we are master
		if !states.IsMaster() {

			//Calculate time
			elapsed := time.Since(heartbeat)
			elapsed = (elapsed + time.Second/2) / time.Second

			if elapsed > 3 {

				//Quit goroutine
				channel_backup_quit <- true

				//Start master
				channel_backup_init_master <- true

				//End goroutine
				fmt.Println(filename, "Failed to receive heartbeat. Closing socket and booting as master.")
				return

			}

		} else {

			//No longer master, quit loop
			break

		}

		time.Sleep(10 * time.Millisecond)

	}
}

/*
	Checks if we received an heartbeat from the backup process locally. Update timestamp on receive.
*/
func HeartbeatListener(channel_backup_init_master chan bool, channel_backup_listen chan utilities.Packet) {

	var message_receive utilities.Message
	var packet_receive utilities.Packet

	//Run loop forever
	for {

		//Check if we are master
		if !states.IsMaster() {

			select {
			case packet_receive = <-channel_backup_listen:

				//Get message and decode
				message_receive = network.DecodeMessage(packet_receive.Data)

				//Checking heartbeat from main process on this PC
				if message_receive.Category == utilities.MESSAGE_HEARTBEAT && packet_receive.Address == network.GetMachineID() {

					setHeartbeat()

				}
			}

		} else {

			//No longer master, quit loop
			break

		}

		time.Sleep(10 * time.Millisecond)

	}

}

/*
	Function sends a message to the local master process each 0.5 seconds.
*/
func HeartbeatLoop(channel_backup_write chan utilities.Packet) {

	//Network messages
	var message_send utilities.Message
	var packet_send utilities.Packet

	for {

		//Check if we are master
		if states.IsMaster() {

			//Create and send message
			message_send.Category = utilities.MESSAGE_HEARTBEAT
			packet_send.Data = network.EncodeMessage(message_send)
			channel_backup_write <- packet_send

		}

		time.Sleep(500 * time.Millisecond)

	}

}

//New heartbeat
func setHeartbeat() {
	heartbeat = time.Now()
}

//Handle exits by Ctrl+C termination.
func ExitHandler() {

	//This channel is called when the program is terminated
	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan

	//Stop the elevator
	driver.Stop()

	//Print
	log.Println(filename, "Program killed! Elevator is stopped.")

	//Do last actions and wait for all write operations to end
	os.Exit(0)
}
