goregen
=======

Golang implementation for controlling regenbox

Install [firmware/firmware.ino](https://github.com/solar3s/goregen/blob/master/firmware/firmware.ino)
on arduino & run code


Quickstart
==========
1/ Download for your OS/Architecture from [builds/](https://github.com/solar3s/goregen/tree/master/builds) directory  
2/ unzip  
3/ upload firmware.ino to Arduino (via Arduino IDE for example)  
4/ run executable  

What to expect ?
  * Serial port should be detected automatically
  * Auto-pilot should be enabled (charge cycles between 900mV & 1400mV)
  * Output to read from stdout in a console

```text
$ ./goregen
2017/04/16 14:16:22 Found serial port "/dev/ttyUSB0"
2017/04/16 14:16:22 Options: &serial.Mode{BaudRate:9600, DataBits:8, Parity:0, StopBits:0}
2017/04/16 14:16:24 enabling discharge
2017/04/16 14:16:25 Voltage: 1228mV
2017/04/16 14:16:26 Voltage: 1226mV
2017/04/16 14:16:27 Voltage: 1223mV
```
