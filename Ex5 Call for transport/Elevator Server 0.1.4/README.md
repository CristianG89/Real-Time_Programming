# Elevator server
In the [TTK4145 elevator project](https://github.com/TTK4145/Project) the elevator hardware is communicated with over TCP. Every elevator expose a server that clients (elevator logic) can connect to. This repo contains the code for the server part of the elevator interface.

## Usage
### Dependencies
#### Hardware & comedi
For this program to work elevator hardware needs to be connected through an io card supported by comedi in the same way as it is in the NTNU real time lab. It is made specially to work with the elevator hardware in this lab, and it is not recommended to use this software for outside this lab. For a solution that will work outside the real time lab, have a look at one of the simulators ([Simulator v2](https://github.com/TTK4145/Simulator-v2), [Simulator v3](https://github.com/TTK4145/Simulator-v3)).

#### Cargo & Rust
To compile and install the elevator server, cargo and rust is needed. Both are best installed by using [rustup](https://www.rustup.rs/). 

### HW Access
For a process to access the io card (elevator hw) on the real time lab the user running the user running the process must be in the iocard group. To add user student to the iocard group run `sudo usermod -a -G iocard student`.

### Install
If the software is not installed you can run `cargo install ttk4145_elevator_server` to run install it. If an old version is installed and you wish to upgrade to the newest version `cargo install --force ttk4145_elevator_server` will do.

### Run
The server can be started by running `ElevatorServer`. Once started, the server will start listening on `localhost:15657`. You can then connect to it by using a [client](https://github.com/TTK4145/elevator-server/new/master?readme=1#clients) that adhers to [the protocol](https://github.com/TTK4145/elevator-server#protocol).

### Clients

## Protocol
 - All TCP messages must have a length of 4 bytes
 - The instructions for reading from the hardware send replies that are 4 bytes long, where the last byte is always 0
 - The instructions for writing to the hardware do not send any replies
 
 ### Write
 
Writing               | `command[0]`         | `command[1]`            | `command[2]`         | `command[3]`
----------------------|----------------------|-------------------------|----------------------|--------------------
Reserved              | x                    | x                       | x                    | x
Motor direction       | 1                    | direction[-1(255),0,1]  | x                    | x
Order button light    | 2                    | button[0,1,2]           | floor<br>[0..NF]     | value[0,1]
Floor indicator       | 3                    | floor[0..NF]            | x                    | x
Door open light       | 4                    | value[0,1]              | x                    | x
Stop button light     | 5                    | value[0,1]              | x                    | x

## Read
Reading               | `command[0]` | `command[1]`  | `command[2]`  | `command[3]` | `response[0]` | `response[1]` | `response[2]` | `response[3]`
----------------------|--------------|---------------|---------------|--------------|---------------|---------------|---------------|---------------
Order button          | 6            | button[0,1,2] | floor[0..NF]  | x            | 6             | active[0,1]   | 0             | 0
Floor sensor          | 7            | x             | x             | x            | 7             | at floor[0,1] | floor[0..NF]  | 0                 
Stop button           | 8            | x             | x             | x            | 8             | active[0,1]   | 0             | 0
Obstruction switch    | 9            | x             | x             | x            | 9             | active[0,1]   | 0             | 0
