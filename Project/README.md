ELEVATOR PROJECT - TTK4145
========================

**Hereafter it is introduced our solution for the term project in lecture TTK4145**: A software with a real time approach to control many lifts (one per computer) throgh libComedi I/O driver, and all of them connected to the same network in the Real Time Lab of NTNU, Trondheim, Norway.
The programming language used is **GO**. The binary file and source files are provided.

Our solution is mainly composed of 6 big modules/packages: **global**, **hw**, **main**, **stateMachine**, **synchroniser** and **network**.
## Modules
### global
The global module basically acts as the header file of the other modules and it has some commonly used constants, variables, structures, channels and functions defined inside.
NOTE: In GO, any global definition from a module/package can be used in another one by: importing the desired module/package, and then writing **owner's module**, **.** (dot) and the **definition** itself (which in addition should start with capital letter).

### hw
The hardware module consists of the **libComedi I/O** driver code that was handed out by professor to work with (in EXERCISE 5). Our solution uses directly the .C files (the most complete version), and this code has been interfaced with a new GO file (this module) so that the lift can be run as well as give feedback to the system.

More specifically, the responsibilities of this module are:
1. Setting/clearing lamps (BUTTONS, FLOOR INDICATOR, STOP, DOOROPEN).
2. Setting motor direction (UP, DOWN, STOP).
3. Reading floor sensor signals.
4. Reading button signals.

### main
As its name says, **this module is the main one, in charge of other modules initialization**. It handles most events in the system (buttons pushed, lamps activation/deactivation, incoming network messages, network reconnection, program termination...).
In addition, main module works as a connection point of the other modules, letting share information amnog them.

### stateMachine
The stateMachine module implements an event-based Finite State Machine (FSM) for the local lift. The lift has 3 states and 3 events.

The 3 possible states of the lift are:
1. **IDLE**: When the lift is at a door in stationary postion with its door closed, awaiting a new order.
2. **MOVING**: The lift is in the state of motion either between floors or at a floor passing it.
3. **DOOR OPEN**: The lift is at a floor with its door open.

The 3 events which can cause the state transitions are:
1. **NEW ORDER**: A new (action) order has been requested. 
2. **FLOOR REACHED**: The lift arrives at any floor.
3. **DOOR TIMEOUT**: The lift door has been open for a while and now it is time to close.

The lift runs based on a queue stored and managed by the synchroniser package.

### synchroniser
The synchroniser module consists of 5 different files: **assigner**, **backup**, **fsmFunctions**, **queue** and **queueFunction**.
Its main task is to take new orders, remove orders and assign orders to the different lifts in a sensible way. It also has some functions to determine if an elevator should stop in a given floor and in which direction it should move.

1. **assigner**'s task is to keep track of which elevator has the lowest cost for the different floors (determined by different factors), and then assign the order to this elevator when order is given.
2. **backup** is there to take a backup of the queue orders and save them to a file whenever a new order is set. It also reads from the file at start-up, to check if there is anything on the queue, in case the program was interrupted before.
3. **fsmFunctions** are functions made for the stateMachine to use. For example choosing motor direction, determine whether it should stop or not and so on.
4. **queue** is the core of this module, where the (local and remote) queue of orders are defined and initialized. It also contains functions like updating the queues, or reassigning a queue if one elevator dies. 
5. **queueFunctions** implements different functions to manage the queues (add, remove, restructure orders).

### network
Having a completely functional local lift, the network module connects it to the network **through UDP communication protocol**, so that all of them can be synchronized. This protocol requires to broadcast all messages through a particular predefined port (one for local and another one for remote). It manages every single UDP incoming and outcoming message, and to be sure all connected lifts are still operative, it broadcasts a periodic life signal message every 400ms.
In case of lost network, this module tries try to reconnect again as soon as possible.

## Assumptions
* Always at least one elevator is operating.
* No network partitioning.
* Stop button and obstructions switch provided are not used for this project.
* If both up and down hall buttons are called on a floor, it is assumed that all people enter the lift when it arrives.
