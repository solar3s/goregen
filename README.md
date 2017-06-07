goregen
=======

For information about the __regenbox__ project or to pre-order hardware, go to https://regenbox.org/

`goregen` is the web remote control interface for your `regenbox`

![goregen](https://cloud.githubusercontent.com/assets/1699009/26520906/429cb2ca-42dc-11e7-948a-8e51deb05e38.png)

installation pre-requisites
---------------------------

This is the boring part, please check the [wiki][1] for detailed instructions  

installation
------------

#### releases

todo

#### go get

`go get -v github.com/solar3s/goregen`

goregen
-------

If you have proper driver, and `goregen`'s firmware was installed to your plugged-in board, you can now run 
it from a terminal and expect something like this:

```
rkj@rkdeb:~/go/src/github.com/solar3s/goregen$ goregen
2017/06/07 23:12:24 using config file: /home/rkj/.goregen/config.toml
2017/06/07 23:12:24 trying "/dev/ttyUSB0"...
2017/06/07 23:12:24 connected to "/dev/ttyUSB0" in 404.795954ms
2017/06/07 23:12:24 starting conn watcher (poll rate: 1s)
2017/06/07 23:12:24 listening on http://localhost:3636 ...

```

#### open web browser at http://localhost:3636/

If you see a blank oscilloscope and the mention `Status: Disconnected`, it means something's not quite ready yet on the
hardware side, please refer to the [wiki][1]

#### goregen -h
```
rkj@rkdeb:~/go/src/github.com/solar3s/goregen$ ./goregen -h
Usage of ./goregen:
  -assets
    	extract static assets to <root>/static, if true, extracted assets also take precedence over binary assets
	this option is useful for doing live tests on front-end
  -config string
    	path to config, defaults to <root>/config.toml
  -debug
    	enable debug mode
  -dev string
    	path to serial port, if empty it will be searched automatically
  -root string
    	path to goregen's config files (default "~/.goregen")
  -verbose
    	higher verbosity
  -version
    	print version & exit
```

[1]: https://github.com/solar3s/goregen/wiki
