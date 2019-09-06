#![allow(dead_code)]

use libc::c_uint;

const COMEDI_INPUT: c_uint = 0;
const COMEDI_OUTPUT: c_uint = 1;
const COMEDI_OPENDRAIN: c_uint = 2;

//in port 4
pub(crate) const PORT_4_SUBDEVICE: c_uint = 3;
pub(crate) const PORT_4_CHANNEL_OFFSET: c_uint = 16;
pub(crate) const PORT_4_DIRECTION: c_uint = COMEDI_INPUT;
pub(crate) const OBSTRUCTION: c_uint = (0x300+23);
pub(crate) const STOP: c_uint = (0x300+22);
pub(crate) const BUTTON_COMMAND0: c_uint = (0x300+21);
pub(crate) const BUTTON_COMMAND1: c_uint = (0x300+20);
pub(crate) const BUTTON_COMMAND2: c_uint = (0x300+19);
pub(crate) const BUTTON_COMMAND3: c_uint = (0x300+18);
pub(crate) const BUTTON_UP0: c_uint = (0x300+17);
pub(crate) const BUTTON_UP1: c_uint = (0x300+16);

//in port 1
pub(crate) const PORT_1_SUBDEVICE: c_uint = 2;
pub(crate) const PORT_1_CHANNEL_OFFSET: c_uint = 0;
pub(crate) const PORT_1_DIRECTION: c_uint = COMEDI_INPUT;
pub(crate) const BUTTON_DOWN1: c_uint = (0x200+0);
pub(crate) const BUTTON_UP2: c_uint = (0x200+1);
pub(crate) const BUTTON_DOWN2: c_uint = (0x200+2);
pub(crate) const BUTTON_DOWN3: c_uint = (0x200+3);
pub(crate) const SENSOR_FLOOR0: c_uint = (0x200+4);
pub(crate) const SENSOR_FLOOR1: c_uint = (0x200+5);
pub(crate) const SENSOR_FLOOR2: c_uint = (0x200+6);
pub(crate) const SENSOR_FLOOR3: c_uint = (0x200+7);

//out port 3
pub(crate) const PORT_3_SUBDEVICE: c_uint = 3;
pub(crate) const PORT_3_CHANNEL_OFFSET: c_uint = 8;
pub(crate) const PORT_3_DIRECTION: c_uint = COMEDI_OUTPUT;
pub(crate) const MOTORDIR: c_uint = 0x300+15;
pub(crate) const LIGHT_STOP: c_uint = (0x300+14);
pub(crate) const LIGHT_COMMAND0: c_uint = (0x300+13);
pub(crate) const LIGHT_COMMAND1: c_uint = (0x300+12);
pub(crate) const LIGHT_COMMAND2: c_uint = (0x300+11);
pub(crate) const LIGHT_COMMAND3: c_uint = (0x300+10);
pub(crate) const LIGHT_UP0: c_uint = (0x300+9);
pub(crate) const LIGHT_UP1: c_uint = (0x300+8);

//out port 2
pub(crate) const PORT_2_SUBDEVICE: c_uint = 3;
pub(crate) const PORT_2_CHANNEL_OFFSET: c_uint = 0;
pub(crate) const PORT_2_DIRECTION: c_uint = COMEDI_OUTPUT;
pub(crate) const LIGHT_DOWN1: c_uint = (0x300+7);
pub(crate) const LIGHT_UP2: c_uint = (0x300+6);
pub(crate) const LIGHT_DOWN2: c_uint = (0x300+5);
pub(crate) const LIGHT_DOWN3: c_uint = (0x300+4);
pub(crate) const LIGHT_DOOR_OPEN: c_uint = (0x300+3);
pub(crate) const LIGHT_FLOOR_IND1: c_uint =(0x300+1);
pub(crate) const LIGHT_FLOOR_IND0: c_uint =(0x300+0);

//out port 0
pub(crate) const MOTOR: c_uint = (0x100+0);
