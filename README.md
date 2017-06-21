goregen
=======

For information about the __regenbox__ project or to pre-order hardware, go to https://regenbox.org/

`goregen` is the web remote control interface for your `regenbox`

![goregen](https://cloud.githubusercontent.com/assets/1699009/26520906/429cb2ca-42dc-11e7-948a-8e51deb05e38.png)

Slack
-----

So far the easiest way to get live help if you're struggling with setup is at https://regenbox.slack.com/
To get auto-invited you need to enter your e-mail address here https://jdid.co/regenbox/slack/

Obviously if you're not struggling you're also very welcome!

Drivers installation
--------------------

Before being able to connect to your Regenbox, you're gonna need hardware drivers.
Refer to the [wiki section][3] for detailed instructions on how to do so.

Installation
------------

#### releases

Download latest release for your OS / architecture at the [releases page][2]

#### via go get

`go get -v github.com/solar3s/goregen`

goregen
-------

If you have proper driver installed, unzip, open a terminal to extracted folder and run `goregen` which should output something like this:

```
rkj@rkdeb:/tmp/goregen$ ./goregen
2017/06/21 12:51:33 created new config file "/tmp/goregen/config.toml"
2017/06/21 12:51:33 using config file: /tmp/goregen/config.toml
2017/06/21 12:51:33 trying "/dev/ttyUSB0"...
2017/06/21 12:51:33 connected to "/dev/ttyUSB0" in 404.825718ms
2017/06/21 12:51:33 starting conn watcher (poll rate: 1s)
2017/06/21 12:51:33 starting webserver on http://localhost:3636 ...
```

configuration
-------------

First run will create a default configuration file `config.toml` in the current directory, it can be modified (restart of `goregen` required)

```
[User]
  Name = ""   # your name for tracking data

[Battery] # the battery you're testing
  Type = ""   # AA, AAA...
  Voltage = 0 # In millivolts
  Brand = ""  # Battery brand
  Model = ""  # Battery model

[Web] # Options related to local web-server
  ListenAddr = "localhost:3636"
  Verbose = false
  StaticDir = ""
  WsInterval = "1s"

[Regenbox] # Regenbox options
  Mode = "Charger"        # Charger / Discharger / Cycler (can be triggered via web-interface)
  NbHalfCycles = 10       # In Cycler mode: number of half-cycles to do before halting auto-mode (0: no-limit holdem)
  UpDuration = "2h0m0s"   # Maximum time for an up-cycle before halting charge 
  DownDuration = "2h0m0s" # Maximum time for a down-cycle before halting discharge
  TopVoltage = 1410       # Target top voltage for charging
  BottomVoltage = 900     # Target bottom voltage for discharging
  IntervalSec = "10s"     # Save data every
  ChargeFirst = true      # In Cycler mode: start with a charge cycle

# Some more advanced configuration below, do not change if not sure

[Watcher]
  ConnPollRate = "1s"     # Check for USB device every

[Serial]
  BaudRate = 57600
  DataBits = 8
  Parity = 0
  StopBits = 0
```

#### open web browser at http://localhost:3636/

If you see a blank oscilloscope and the mention `Status: Disconnected`, it means something's not quite ready yet on the
hardware side, please refer to the [wiki][1]

#### goregen -h
```
rkj@rkdeb:~/go/src/github.com/solar3s/goregen$ ./goregen -h
Usage of ./goregen:
  -assets
    	extract static assets to <root>/static, if true, extracted assets also take precedence over binary assets. This option is useful for doing live tests on front-end
  -config string
    	path to config (defaults to <root>/config.toml)
  -dev string
    	path to serial port, if empty it will be searched automatically
  -root string
    	path to goregen's main directory (defaults to executable path)
  -verbose
    	higher verbosity
  -version
    	print version & exit
```

Contributing
------------

The project is oh-so-young and needs help from everyone.
Testing and documenting comes to mind, in the [wiki section][1] for example which needs love, feedback and evolution with the project.
Any form of contribution to the code-base is also definitely welcome, for now no specifics are in place but usual guidelines apply. In any case, slack.

#### Bug reports
Should you encounter a bug, please provide as much information as possible and post an issue if it's not already there : https://github.com/solar3s/goregen/issues - if you're not sure, slack.

[1]: https://github.com/solar3s/goregen/wiki
[2]: https://github.com/solar3s/goregen/releases
[3]: https://github.com/solar3s/goregen/wiki/Driver-installation
