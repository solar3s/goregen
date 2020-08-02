goregen
=======

For information about the __regenbox__ project or to pre-order hardware, go to https://regenbox.org/

`goregen` is the web remote control interface for your `regenbox`

![goregen](https://chatmetaleux.be/multi_batteries_0.5.0.png)

Version X4
----------

This version will work only for your regenbox if you wired it with the 4 batteries pins to the arduino.
See more here : https://regenbox.slack.com/archives/C4V8L17D3/p1594030190006400

The firmware has been updated to handle reading 4 voltages pins.

The data logger will always log 4 voltages curves. So it can work for an official test if you only take the battery #1.

Some variables names like vVoltage, vVoltage1, config.Battery1... need to be refactored for better understanding.

Marshaller has been updated, but I'm not sure this will works well eveywhere.

The cmd arg -assets won't work as the go-bindata util changed.

TODO : Need to add code check to server.go to handle zero battery charging.

Slack
-----

We've got a slack running at https://regenbox.slack.com/  
To get auto-invited you need to enter your e-mail address here https://jdid.co/regenbox/slack/

Everyone's welcome to chat about the project or get live help from the team.

Drivers installation
--------------------

Before being able to connect to your Regenbox, you're gonna need arduino drivers.  
Refer to the [drivers section][3] for detailed instructions on how to do so.

Installation
------------

#### releases page

Download latest release for your OS / architecture at the [releases page][2]

#### or via go get

You can download from source using `go get`, if you're familiar with it.  
`go get -v github.com/solar3s/goregen`

Firmware installation
---------------------

After drivers are installed and you downloaded your `goregen` release, the next step 
is to upload the file `firmware.ino` into the Arduino of your Regenbox.  
Once again refer to [firmware section][4] for detailed instructions. 

goregen
-------

If you have proper driver installed & the firmware installed, open a terminal to the extracted folder 
and run `goregen` which should output something like this:

```
rkj@rkdeb:/tmp/goregen$ ./goregen
2017/06/21 12:51:33 using config file: /tmp/goregen/config.toml
2017/06/21 12:51:33 trying "/dev/ttyUSB0"...
2017/06/21 12:51:33 connected to "/dev/ttyUSB0" in 404.825718ms
2017/06/21 12:51:33 starting conn watcher (poll rate: 1s)
2017/06/21 12:51:33 starting webserver on http://localhost:3636 ...
```

configuration
-------------

A default configuration `config.toml` is shipped with `goregen` releases.  
You need to modify it using a text editor with appropriate values where needed, restart of `goregen` is required 
for reloading configuration.

```
Resistor = "10.2"               # The value of the resistor you're using in Ohms
# Device = "/dev/myRegenBox"    # By default goregen will try to detect your regenbox, no need to specify this (advanced use)

[User]
  BetaId = ""                   # Beta ID as provided with your Regenbox
  Name = ""                     # Your aka

[Battery]
  BetaRef = ""                  # Beta reference as provided with your Regenbox
  Type = ""                     # AAA / AA...
  Voltage = 0                   # In millivolts
  Brand = ""                    # Battery brand
  Model = ""                    # Battery model

[Regenbox]
  Mode = "Charger"              # Charger / Discharger / Cycler (usually set using web interface controls)
  NbHalfCycles = 10             # In Cycler mode, number of half-cycles to do before stopping 
  UpDuration = "24h0m0s"        # Maximum duration for a Charge cycle
  DownDuration = "24h0m0s"      # Maximum duration for a Discharge cycle
  TopVoltage = 1500             # Target voltage for a Charge cycle 
  BottomVoltage = 900           # Target voltage for a Discharge cycle
  Ticker = "10s"                # Check & save to datalog battery voltage every
  ChargeFirst = false           # In Cycler mode, start with a charge cycle if true, else discharge
  
#------------These are a bit more advanced and shouldn't need modification. 

[Web]
  ListenAddr = "localhost:3636" # Listening address & port for the local server
  StaticDir = "static"          # Path to static assets, extracted with goregen -assets
  DataDir = "data"              # Path to charts datalogs
  WebsocketInterval = "1s"      # Check regenbox state on web interface every

[Watcher]
  ConnPollRate = "1s"           # Check USB connection to Regenbox every

[Serial]                        # Advanced serial settings
  BaudRate = 57600
  DataBits = 8
  Parity = 0
  StopBits = 0
```

#### open web browser at http://localhost:3636/

If you see a blank oscilloscope and the mention `Status: Disconnected`, it means something's not quite ready yet on the
hardware side, please refer to the [drivers section][1]

#### goregen -h
```
rkj@rkdeb:~/go/src/github.com/solar3s/goregen$ ./goregen -h
Usage of ./goregen:
  -assets string
    	restore static assets to provided directory & exit
  -config string
    	path to config (defaults to <root>/config.toml)
  -data string
    	path to data directory (defaults to <root>/data)
  -dev string
    	path to serial port, if empty it will be searched automatically
  -log string
    	path to logs directory (defaults to <root>/log)
  -root string
    	path to goregen's main directory (defaults to executable path)
  -v	higher verbosity
  -version
    	print version & exit
```

Contributing
------------

The project is oh-so-young and needs help from everyone.
Testing and documenting comes to mind, in the [wiki section][1] for example which needs love, feedback and evolution with the project.
Any form of contribution to the code-base is also definitely welcome, for now no specifics are in place but usual guidelines apply.

#### Bug reports

Please provide as much information as possible and post an issue at the issues page : https://github.com/solar3s/goregen/issues

[1]: https://github.com/solar3s/goregen/wiki
[2]: https://github.com/solar3s/goregen/releases
[3]: https://github.com/solar3s/goregen/wiki/Driver-installation
[4]: https://github.com/solar3s/goregen/wiki/Upgrading-firmware
