// Defines interactions with lift HW (libComedi I/O) at real time lab
// at Department of Engineering Cybernetics at NTNU, Trondheim, Norway.

package hw

import (
	"../global"
	"errors"
	"log"
)

// Matrix for the lights
var lampChannelMatrix = [global.NumFloors][global.NumButtons]int{
	{LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
	{LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
	{LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
	{LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
}
// Matrix for the buttons
var buttonChannelMatrix = [global.NumFloors][global.NumButtons]int{
	{BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
	{BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
	{BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
	{BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
}

// Initialises the lift HW and moves lift to a defined state (descends until it reaches a floor)
func HW_Init() (int, error) {
	if !ioInit() {
		return -1, errors.New("Hardware driver: ioInit() failed!")
	}

	// Switch off all floor button lamps
	for f := 0; f < global.NumFloors; f++ {
		if f != 0 {
			SetButtonLamp(f, global.BtnOutDown, false)
		}
		if f != global.NumFloors-1 {
			SetButtonLamp(f, global.BtnOutUp, false)
		}
		SetButtonLamp(f, global.BtnInside, false)
	}
	// Switch off the door open and the stop buttons
	SetStopLamp(false)
	SetDoorLamp(false)

	// Move to defined state and switch on the corresponding floor lamp
	SetMotorDir(global.DirDown)
	floor := Floor()
	for floor == -1 {
		floor = Floor()
	}
	SetMotorDir(global.DirStop)
	SetFloorLamp(floor)

	log.Println(global.ColorG, "Hardware initialised.", global.ColorN)
	return floor, nil
}

// Sets the motor direction to move up, move down or stop.
func SetMotorDir(dirn int) {
	if dirn == 0 {
		ioWriteAnalog(MOTOR, 0)
	} else if dirn > 0 {
		ioClearBit(MOTORDIR)
		ioWriteAnalog(MOTOR, 2800)
	} else if dirn < 0 {
		ioSetBit(MOTORDIR)
		ioWriteAnalog(MOTOR, 2800)
	}
}

// Checks for invalid button presses and returns true when right one is pressed
func ReadButton(floor int, button int) bool {
	if floor < 0 || floor >= global.NumFloors {
		log.Println("Error: Floor %d out of range!\n", floor)
		return false
	}
	if button < 0 || button >= global.NumButtons {
		log.Println("Error: Button %d out of range!\n", button)
		return false
	}
	if button == global.BtnOutUp && floor == global.NumFloors-1 {
		log.Println("Button up from top floor does not exist!")
		return false
	}
	if button == global.BtnOutDown && floor == 0 {
		log.Println("Button down from ground floor does not exist!")
		return false
	}
	if ioReadBit(buttonChannelMatrix[floor][button]) {
		return true
	} else {
		return false
	}
}

// Checks for invalid button presses and switches on when right button is pressed
func SetButtonLamp(floor int, button int, value bool) {
	if floor < 0 || floor >= global.NumFloors {
		log.Println("Error: Floor %d out of range!\n", floor)
		return
	}
	if button == global.BtnOutUp && floor == global.NumFloors-1 {
		log.Println("Button up from top floor does not exist!")
		return
	}
	if button == global.BtnOutDown && floor == 0 {
		log.Println("Button down from ground floor does not exist!")
		return
	}
	if button != global.BtnOutUp &&
		button != global.BtnOutDown &&
		button != global.BtnInside {
		log.Println("Invalid button %d\n", button)
		return
	}
	if value {
		ioSetBit(lampChannelMatrix[floor][button])
	} else {
		ioClearBit(lampChannelMatrix[floor][button])
	}
}

// Senses which floor the lift is in and returns an integer
func Floor() int {
	if ioReadBit(SENSOR_FLOOR1) {
		return 0
	} else if ioReadBit(SENSOR_FLOOR2) {
		return 1
	} else if ioReadBit(SENSOR_FLOOR3) {
		return 2
	} else if ioReadBit(SENSOR_FLOOR4) {
		return 3
	} else {
		return -1
	}
}

// Switches on/off the floor buttons 
func SetFloorLamp(floor int) {
	if floor < 0 || floor >= global.NumFloors {
		log.Println("Error: Floor %d out of range!\n", floor)
		log.Println("No floor indicator will be set.")
		return
	}

	// Binary encoding. One light must always be on.
	if floor&0x02 > 0 {
		ioSetBit(LIGHT_FLOOR_IND1)
	} else {
		ioClearBit(LIGHT_FLOOR_IND1)
	}

	if floor&0x01 > 0 {
		ioSetBit(LIGHT_FLOOR_IND2)
	} else {
		ioClearBit(LIGHT_FLOOR_IND2)
	}
}

// Switches on/off the door lamp by setting the bit or clearing it
func SetDoorLamp(value bool) {
	if value {
		ioSetBit(LIGHT_DOOR_OPEN)
	} else {
		ioClearBit(LIGHT_DOOR_OPEN)
	}
}

// Switches on/off the stop lamp (not used in this project!)
func SetStopLamp(value bool) {
	if value {
		ioSetBit(LIGHT_STOP)
	} else {
		ioClearBit(LIGHT_STOP)
	}
}

// Returns true when an obstruction/stop is read (not used in this project!)
func GetObstructionSignal() bool {
	return ioReadBit(OBSTRUCTION)
}

// Reads stop buuton (not used in this project!)
func GetStopSignal() bool {
	return ioReadBit(STOP)
}