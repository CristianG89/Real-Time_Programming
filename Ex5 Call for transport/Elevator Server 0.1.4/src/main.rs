extern crate libc;

#[cfg(test)]
#[macro_use]
extern crate lazy_static;
#[cfg(test)]
extern crate rand;

use std::io::{Read, Write};

use std::net::TcpListener;
use std::ffi::CString;

mod channel;

/// This is a opaque rust equivalent for comedi_t inside libcomedi.h
#[allow(non_camel_case_types)]
enum comedi_t {}

#[link(name = "comedi")]
extern "C" {
    fn comedi_open(interface_name: *const libc::c_char) -> *const comedi_t;
    fn comedi_dio_config(it: *const comedi_t, subd: libc::c_uint, chan: libc::c_uint, dir: libc::c_uint) -> libc::c_int;
    fn comedi_dio_write(it: *const comedi_t, subd: libc::c_uint, chan: libc::c_uint, bit: libc::c_uint) -> libc::c_int;
    fn comedi_dio_read(it: *const comedi_t, subd: libc::c_uint, chan: libc::c_uint, bit: *mut libc::c_uint) -> libc::c_int;
    fn comedi_data_write(it: *const comedi_t, subd: libc::c_uint, chan: libc::c_uint, range: libc::c_uint, aref: libc::c_uint, data: libc::c_uint) -> libc::c_int;
}

enum Command {
    Reserved,
    WriteMotorDirection(ElevatorDirection),
    WriteOrderButtonLight(ButtonType, u8, bool),
    WriteFloorIndicator(u8),
    WriteDoorOpenLight(bool),
    WriteStopButtonLight(bool),
    ReadOrderButton(ButtonType, u8),
    ReadFloorSensor,
    ReadStopButton,
    ReadObstructionSwitch,
}

impl Command {
    fn decode(data: &[u8]) -> Self {
        assert_eq!(data.len(), 4);
        match data[0] {
            0 => Command::Reserved,
            1 => Command::WriteMotorDirection(ElevatorDirection::decode(data[1])),
            2 => Command::WriteOrderButtonLight(ButtonType::decode(data[1]), data[2], data[3] != 0),
            3 => Command::WriteFloorIndicator(data[1]),
            4 => Command::WriteDoorOpenLight(data[1] != 0),
            5 => Command::WriteStopButtonLight(data[1] != 0),
            6 => Command::ReadOrderButton(ButtonType::decode(data[1]), data[2]),
            7 => Command::ReadFloorSensor,
            8 => Command::ReadStopButton,
            9 => Command::ReadObstructionSwitch,
            x => panic!("Not a valid command code: {}", x),
        }
    }
}

pub enum ElevatorDirection{
    Up,
    Down,
    Stop,
}

impl ElevatorDirection {
    fn decode(data: u8) -> Self {
        match data {
            0 => ElevatorDirection::Stop,
            1 => ElevatorDirection::Up,
            255 => ElevatorDirection::Down,
            x => panic!("Not a valid direction code: {}", x),
        }
    }
}

#[derive(Debug, PartialEq, Clone, Copy)]
enum ButtonType {
    HallUp,
    HallDown,
    Cab,
}

impl ButtonType {
    fn decode(data: u8) -> Self {
        match data {
            0 => ButtonType::HallUp,
            1 => ButtonType::HallDown,
            2 => ButtonType::Cab,
            x => panic!("Not a valid ButtonType code: {}", x),
        }
    }
}


pub struct ElevatorInterface(*const comedi_t);

unsafe impl Send for ElevatorInterface {}

impl ElevatorInterface {
    const MOTOR_SPEED: u32 = 2800;
    const N_FLOORS: u8 = 4;
    
    fn open(interface_name: &str) -> Result<Self, ()> {
        unsafe {
            let comedi = comedi_open(CString::new(interface_name).unwrap().as_ptr());
            
            if comedi.is_null() {
                Err(())
            } else {
	        let mut status = 0;
                for i in 0..8 {
                    status |= comedi_dio_config(comedi, channel::PORT_1_SUBDEVICE, i, channel::PORT_1_DIRECTION);
                    status |= comedi_dio_config(comedi, channel::PORT_2_SUBDEVICE, i, channel::PORT_2_DIRECTION);
                    status |= comedi_dio_config(comedi, channel::PORT_3_SUBDEVICE, i+8, channel::PORT_3_DIRECTION);
                    status |= comedi_dio_config(comedi, channel::PORT_4_SUBDEVICE, i+16, channel::PORT_4_DIRECTION);
                }
                
		if status == 0 {
		    Ok(ElevatorInterface(comedi))
		} else {
		    Err(())
                }
            }
        }
    }
    
    fn set_direction(&self, dir: ElevatorDirection) {
        unsafe {
            match dir {
                ElevatorDirection::Up => {
                    comedi_dio_write(self.0, channel::MOTORDIR >> 8, channel::MOTORDIR & 0xff, 0);
                    comedi_data_write(self.0, channel::MOTOR >> 8, channel::MOTOR & 0xff, 0, 0, Self::MOTOR_SPEED);
                },
                ElevatorDirection::Down => {
                    comedi_dio_write(self.0, channel::MOTORDIR >> 8, channel::MOTORDIR & 0xff, 1);
                    comedi_data_write(self.0, channel::MOTOR >> 8, channel::MOTOR & 0xff, 0, 0, Self::MOTOR_SPEED);
                },
                ElevatorDirection::Stop => {
                    comedi_data_write(self.0, channel::MOTOR >> 8, channel::MOTOR & 0xff, 0, 0, 0);
                },
            }
        }
    }

    fn read_floor_sensor(&self) -> Option<u8> {
        unsafe {
            let mut data: libc::c_uint = 0;
            comedi_dio_read(self.0, channel::SENSOR_FLOOR0 >> 8, channel::SENSOR_FLOOR0 & 0xff, &mut data);
            if data != 0 {
                return Some(0);
            }
            
            comedi_dio_read(self.0, channel::SENSOR_FLOOR1 >> 8, channel::SENSOR_FLOOR1 & 0xff, &mut data);
            if data != 0 {
                return Some(1);
            }
            
            comedi_dio_read(self.0, channel::SENSOR_FLOOR2 >> 8, channel::SENSOR_FLOOR2 & 0xff, &mut data);
            if data != 0 {
                return Some(2);
            }
            
            comedi_dio_read(self.0, channel::SENSOR_FLOOR3 >> 8, channel::SENSOR_FLOOR3 & 0xff, &mut data);
            if data != 0 {
                return Some(3);
            }
            
            None
        }
    }

    fn set_order_button_light(&self, button_type: ButtonType, floor: u8, on_not_off: bool) {
        assert!(floor < ElevatorInterface::N_FLOORS);
        unsafe {
            match (button_type, floor) {
                (ButtonType::HallUp, 0) => comedi_dio_write(self.0, channel::LIGHT_UP0 >> 8, channel::LIGHT_UP0 & 0xff, on_not_off as libc::c_uint),
                (ButtonType::Cab, 0) => comedi_dio_write(self.0, channel::LIGHT_COMMAND0 >> 8, channel::LIGHT_COMMAND0 & 0xff, on_not_off as libc::c_uint),
		(ButtonType::HallDown, 0) => 0,
                (ButtonType::HallUp, 1) => comedi_dio_write(self.0, channel::LIGHT_UP1 >> 8, channel::LIGHT_UP1 & 0xff, on_not_off as libc::c_uint),
                (ButtonType::HallDown, 1) => comedi_dio_write(self.0, channel::LIGHT_DOWN1 >> 8, channel::LIGHT_DOWN1 & 0xff, on_not_off as libc::c_uint),
                (ButtonType::Cab, 1) => comedi_dio_write(self.0, channel::LIGHT_COMMAND1 >> 8, channel::LIGHT_COMMAND1 & 0xff, on_not_off as libc::c_uint),
                (ButtonType::HallUp, 2) => comedi_dio_write(self.0, channel::LIGHT_UP2 >> 8, channel::LIGHT_UP2 & 0xff, on_not_off as libc::c_uint),
                (ButtonType::HallDown, 2) => comedi_dio_write(self.0, channel::LIGHT_DOWN2 >> 8, channel::LIGHT_DOWN2 & 0xff, on_not_off as libc::c_uint),
                (ButtonType::Cab, 2) => comedi_dio_write(self.0, channel::LIGHT_COMMAND2 >> 8, channel::LIGHT_COMMAND2 & 0xff, on_not_off as libc::c_uint),
		(ButtonType::HallUp, 3) => 0,
                (ButtonType::HallDown, 3) => comedi_dio_write(self.0, channel::LIGHT_DOWN3 >> 8, channel::LIGHT_DOWN3 & 0xff, on_not_off as libc::c_uint),
                (ButtonType::Cab, 3) => comedi_dio_write(self.0, channel::LIGHT_COMMAND3 >> 8, channel::LIGHT_COMMAND3 & 0xff, on_not_off as libc::c_uint),
                (b, f) => panic!("You tried to set lamp in non-existing button: {:?}:{} <button:floor>", b, f), //TODO: implement display for ButtonType
            };
        }
    }

    fn read_order_button(&self, button_type: ButtonType, floor: u8) -> bool {
        assert!(floor < 4);
        unsafe {
            let mut data: libc::c_uint = 0;
            match (button_type, floor) {
                (ButtonType::HallUp, 0) => comedi_dio_read(self.0, channel::BUTTON_UP0 >> 8, channel::BUTTON_UP0 & 0xff, &mut data),
                (ButtonType::Cab, 0) => comedi_dio_read(self.0, channel::BUTTON_COMMAND0 >> 8, channel::BUTTON_COMMAND0 & 0xff, &mut data),
		(ButtonType::HallDown, 0) => 0,
                (ButtonType::HallUp, 1) => comedi_dio_read(self.0, channel::BUTTON_UP1 >> 8, channel::BUTTON_UP1 & 0xff, &mut data),
                (ButtonType::HallDown, 1) => comedi_dio_read(self.0, channel::BUTTON_DOWN1 >> 8, channel::BUTTON_DOWN1 & 0xff, &mut data),
                (ButtonType::Cab, 1) => comedi_dio_read(self.0, channel::BUTTON_COMMAND1 >> 8, channel::BUTTON_COMMAND1 & 0xff, &mut data),
                (ButtonType::HallUp, 2) => comedi_dio_read(self.0, channel::BUTTON_UP2 >> 8, channel::BUTTON_UP2 & 0xff, &mut data),
                (ButtonType::HallDown, 2) => comedi_dio_read(self.0, channel::BUTTON_DOWN2 >> 8, channel::BUTTON_DOWN2 & 0xff, &mut data),
                (ButtonType::Cab, 2) => comedi_dio_read(self.0, channel::BUTTON_COMMAND2 >> 8, channel::BUTTON_COMMAND2 & 0xff, &mut data),
		(ButtonType::HallUp, 3) => 0,
                (ButtonType::HallDown, 3) => comedi_dio_read(self.0, channel::BUTTON_DOWN3 >> 8, channel::BUTTON_DOWN3 & 0xff, &mut data),
                (ButtonType::Cab, 3) => comedi_dio_read(self.0, channel::BUTTON_COMMAND3 >> 8, channel::BUTTON_COMMAND3 & 0xff, &mut data),
                (b, f) => panic!("You tried to set lamp in non-existing button: {:?}:{} <button:floor>", b, f), //TODO: implement display for ButtonType
            };
            data != 0
        }
    }

    fn set_stop_button_light(&self, on_not_off: bool) {
        unsafe {
            comedi_dio_write(self.0, channel::LIGHT_STOP >> 8, channel::LIGHT_STOP & 0xff, on_not_off as libc::c_uint);
        }
    }

    fn read_stop_button(&self) -> bool {
        unsafe{
            let mut data: libc::c_uint = 0;
            comedi_dio_read(self.0, channel::STOP >> 8, channel::STOP & 0xff, &mut data);
            data != 0
        }
    }

    fn set_floor_indicator(&self, floor: u8) {
        assert!(floor < 4);
        unsafe {
            comedi_dio_write(self.0, channel::LIGHT_FLOOR_IND0 >> 8, channel::LIGHT_FLOOR_IND0 & 0xff, ((floor & 1<<1) != 0) as u32);
            comedi_dio_write(self.0, channel::LIGHT_FLOOR_IND1 >> 8, channel::LIGHT_FLOOR_IND1 & 0xff, ((floor & 1<<0) != 0) as u32);
        }
    }

    fn set_door_light(&self, on_not_off: bool) {
        unsafe {
            comedi_dio_write(self.0, channel::LIGHT_DOOR_OPEN >> 8, channel::LIGHT_DOOR_OPEN & 0xff, on_not_off as libc::c_uint);
        }
    }

    fn read_obstruction_sensor(&self) -> bool {
        unsafe {
            let mut data: libc::c_uint = 0;
            comedi_dio_read(self.0, channel::OBSTRUCTION >> 8, channel::OBSTRUCTION & 0xff, &mut data);
            data != 0
        }
    }
}

impl Drop for ElevatorInterface {
    fn drop(&mut self) {
        self.set_direction(ElevatorDirection::Stop);
    }
}


fn main() {
    println!("Elevator server started");
    let (mut stream, _addr) = TcpListener::bind("localhost:15657").unwrap().accept().unwrap();
    let elevator = ElevatorInterface::open("/dev/comedi0").unwrap();
    println!("Client connected to server");
    
    loop {
        let mut received_data = [0u8; 4];
        if let Err(_) = stream.read_exact(&mut received_data) {
            println!("Lost connection to client");
            return;
        }
        let command = Command::decode(&received_data);
        
        match command {
            Command::Reserved => (),
            Command::WriteMotorDirection(dir) => elevator.set_direction(dir),
            Command::WriteOrderButtonLight(button, floor, state) => elevator.set_order_button_light(button, floor, state),
            Command::WriteFloorIndicator(floor) => elevator.set_floor_indicator(floor),
            Command::WriteDoorOpenLight(state) => elevator.set_door_light(state),
            Command::WriteStopButtonLight(state) => elevator.set_stop_button_light(state),
            Command::ReadOrderButton(button, floor) => {
                let response_data = [6u8, elevator.read_order_button(button, floor) as u8, 0, 0];
                stream.write_all(&response_data).unwrap();
            },
            Command::ReadFloorSensor => {
                let response_data = match elevator.read_floor_sensor() {
                    Some(floor) => [7u8, 1, floor, 0],
                    None => [7u8, 0, 0, 0],
                };
                stream.write_all(&response_data).unwrap();
            },
            Command::ReadStopButton => {
                let response_data = [9u8, elevator.read_stop_button() as u8, 0, 0];
                stream.write_all(&response_data).unwrap();
            },
            Command::ReadObstructionSwitch => {
                let response_data = [9u8, elevator.read_obstruction_sensor() as u8, 0, 0];
                stream.write_all(&response_data).unwrap();
            },

        }

    }
}

#[cfg(test)]
mod tests {
    use *;

    use std::sync::Mutex;
    use std::thread;
    use std::time::Duration;

    // These tests are executed on an actual elevator. To make sure only one test is run at the same time, the elevator is protected by this mutex.
    lazy_static! {
        static ref ELEVATOR: Mutex<ElevatorInterface> = {
            let elevator = ElevatorInterface::open("/dev/comedi0").unwrap();

            for f in 0..ElevatorInterface::N_FLOORS { elevator.set_order_button_light(ButtonType::Cab, f, false); }
            for f in 1..ElevatorInterface::N_FLOORS { elevator.set_order_button_light(ButtonType::HallDown, f, false); }
            for f in 0..ElevatorInterface::N_FLOORS-1 { elevator.set_order_button_light(ButtonType::HallUp, f, false); }
            elevator.set_stop_button_light(false);
            
            Mutex::new(elevator)
        };
    }
    
    
    #[test]
    fn init_elevator() {
        ELEVATOR.lock().unwrap();
    }

    #[test]
    fn test_run() {
        let elevator = ELEVATOR.lock().unwrap();
        println!("The elevator will now do a run from the bottom floor to the top floor. It will stop in the floor below the top floor");
        elevator.set_direction(ElevatorDirection::Down);
        while elevator.read_floor_sensor() != Some(0) {
            if let Some(floor) = elevator.read_floor_sensor() {
                elevator.set_floor_indicator(floor);
            }
        }
        elevator.set_floor_indicator(0);
        elevator.set_direction(ElevatorDirection::Up);
        while elevator.read_floor_sensor() != Some(ElevatorInterface::N_FLOORS-1) {
            if let Some(floor) = elevator.read_floor_sensor() {
                elevator.set_floor_indicator(floor);
            }
        }
        elevator.set_floor_indicator(ElevatorInterface::N_FLOORS-1);
        elevator.set_direction(ElevatorDirection::Down);
        while elevator.read_floor_sensor() != Some(ElevatorInterface::N_FLOORS-2) {}
        elevator.set_floor_indicator(ElevatorInterface::N_FLOORS-2);
        elevator.set_direction(ElevatorDirection::Stop);
    }
    
    #[test]
    fn test_cab_buttons() {
        let elevator = ELEVATOR.lock().unwrap();

        for i in rand::seq::sample_indices(&mut rand::thread_rng(), ElevatorInterface::N_FLOORS as usize, ElevatorInterface::N_FLOORS as usize).into_iter() {
            elevator.set_order_button_light(ButtonType::Cab, i as u8, true);
            thread::sleep(Duration::new(0, 200000000));
            elevator.set_order_button_light(ButtonType::Cab, i as u8, false);
            thread::sleep(Duration::new(0, 200000000));
            elevator.set_order_button_light(ButtonType::Cab, i as u8, true);
            while !elevator.read_order_button(ButtonType::Cab, i as u8) {}
            elevator.set_order_button_light(ButtonType::Cab, i as u8, false);
        }
    }

    #[test]
    fn test_hall_up_buttons() {
        let elevator = ELEVATOR.lock().unwrap();

        for i in rand::seq::sample_indices(&mut rand::thread_rng(), ElevatorInterface::N_FLOORS as usize - 1, ElevatorInterface::N_FLOORS as usize - 1).into_iter() {
            elevator.set_order_button_light(ButtonType::HallUp, i as u8, true);
            thread::sleep(Duration::new(0, 200000000));
            elevator.set_order_button_light(ButtonType::HallUp, i as u8, false);
            thread::sleep(Duration::new(0, 200000000));
            elevator.set_order_button_light(ButtonType::HallUp, i as u8, true);
            while !elevator.read_order_button(ButtonType::HallUp, i as u8) {}
            elevator.set_order_button_light(ButtonType::HallUp, i as u8, false);
        }
    }

    #[test]
    fn test_hall_down_buttons() {
        let elevator = ELEVATOR.lock().unwrap();

        for i in rand::seq::sample_indices(&mut rand::thread_rng(), ElevatorInterface::N_FLOORS as usize - 1, ElevatorInterface::N_FLOORS as usize - 1).into_iter() {
            elevator.set_order_button_light(ButtonType::HallDown, i as u8 + 1, true);
            thread::sleep(Duration::new(0, 200000000));
            elevator.set_order_button_light(ButtonType::HallDown, i as u8 + 1, false);
            thread::sleep(Duration::new(0, 200000000));
            elevator.set_order_button_light(ButtonType::HallDown, i as u8 + 1, true);
            while !elevator.read_order_button(ButtonType::HallDown, i as u8 + 1) {}
            elevator.set_order_button_light(ButtonType::HallDown, i as u8 + 1, false);
        }
    }

    #[test]
    fn test_stop_button() {
        let elevator = ELEVATOR.lock().unwrap();
        
        elevator.set_stop_button_light(true);
        thread::sleep(Duration::new(0, 200000000));
        elevator.set_stop_button_light(false);
        thread::sleep(Duration::new(0, 200000000));
        elevator.set_stop_button_light(true);
        while !elevator.read_stop_button() {}
        elevator.set_stop_button_light(false);
    }

    #[test]
    fn door_test() {
        let elevator = ELEVATOR.lock().unwrap();

        for i in 0..4 {
            elevator.set_door_light(true);
            thread::sleep(Duration::new(0, 100000000));
            elevator.set_door_light(false);
            thread::sleep(Duration::new(0, 100000000));
        }
        
        elevator.set_door_light(true);
        while !elevator.read_obstruction_sensor() {}
        thread::sleep(Duration::new(0, 500000000));
        elevator.set_door_light(false);
        while elevator.read_obstruction_sensor() {}
        thread::sleep(Duration::new(0, 500000000));
        elevator.set_door_light(true);
        while !elevator.read_obstruction_sensor() {}
        thread::sleep(Duration::new(0, 500000000));
        elevator.set_door_light(false);
        while elevator.read_obstruction_sensor() {}
        thread::sleep(Duration::new(0, 500000000));
        elevator.set_door_light(true);
        while !elevator.read_obstruction_sensor() {}
        thread::sleep(Duration::new(0, 500000000));
        elevator.set_door_light(false);
        while elevator.read_obstruction_sensor() {}
    }

}
