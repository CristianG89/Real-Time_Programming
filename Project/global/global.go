package global

import (
	"os/exec"
)

// Colours for printing text on Terminal
const Color0 = "\x1b[30;1m" // Dark grey
const ColorR = "\x1b[31;1m" // Red
const ColorG = "\x1b[32;1m" // Green
const ColorY = "\x1b[33;1m" // Yellow
const ColorB = "\x1b[34;1m" // Blue
const ColorM = "\x1b[35;1m" // Magenta
const ColorC = "\x1b[36;1m" // Cyan
const ColorW = "\x1b[37;1m" // White
const ColorN = "\x1b[0m"    // Grey

// Global constants
const NumButtons = 3
const NumFloors = 4

// Local IP address
var Laddr string

var SyncLightsChan = make(chan bool)
var CloseConnectionChan = make(chan bool)

// Opens a new terminal with command "main" when Restart.Run()
var Restart = exec.Command("gnome-terminal", "-x", "sh", "-c", "main")

const (		//Buttons constants
	BtnOutUp int = iota		//Starting from 0, every label acquires a ordered number
	BtnOutDown				//(equivalent to "enum" in C)
	BtnInside
)

const (		//Motor direction constants
	DirDown int = iota - 1
	DirStop
	DirUp
)

const (		//Network message type constants
	Msg_Alive int = iota + 1
	Msg_NewOrder
	Msg_CompleteOrder
	Msg_Cost
)

type FSM_channels struct {
	// Channels for events occurence
	NewOrder     chan bool
	FloorReached chan int
	DoorTimeout  chan bool
	// Channels for Hardware interaction
	MotorDir  chan int
	FloorLamp chan int
	DoorLamp  chan bool
	// Channel for the Door timer
	DoorTimerReset chan bool
	// Channel for interaction with Network 
	OutgoingMsg chan Message
}

type Keypress struct {
	Button int
	Floor  int
}

// Generic network message. No other messages are ever sent on the network.
type Message struct {
	Category int
	Floor    int
	Button   int
	Cost     int
	Addr     string `json:"-"`
}