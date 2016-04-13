package driver

//Imports
import (
	"../utilities"
	"../hardware"
)

//Variables
var button_matrix[utilities.FLOORS][utilities.BUTTONS] int = [utilities.FLOORS][utilities.BUTTONS] int{
	{hardware.BUTTON_UP1, hardware.BUTTON_DOWN1, hardware.BUTTON_COMMAND1},
    {hardware.BUTTON_UP2, hardware.BUTTON_DOWN2, hardware.BUTTON_COMMAND2},
    {hardware.BUTTON_UP3, hardware.BUTTON_DOWN3, hardware.BUTTON_COMMAND3},
    {hardware.BUTTON_UP4, hardware.BUTTON_DOWN4, hardware.BUTTON_COMMAND4},
} 

//Variables
var lamp_matrix[utilities.FLOORS][utilities.BUTTONS] int = [utilities.FLOORS][utilities.BUTTONS] int{
    {hardware.LIGHT_UP1, hardware.LIGHT_DOWN1, hardware.LIGHT_COMMAND1},
    {hardware.LIGHT_UP2, hardware.LIGHT_DOWN2, hardware.LIGHT_COMMAND2},
    {hardware.LIGHT_UP3, hardware.LIGHT_DOWN3, hardware.LIGHT_COMMAND3},
    {hardware.LIGHT_UP4, hardware.LIGHT_DOWN4, hardware.LIGHT_COMMAND4}, 
}