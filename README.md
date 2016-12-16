# Under implementing!

so interface and API will be changed

# sensorbee-iotop

**memo**

path: `$GOPATH/src/github.com/sensorbee/sensorbee-iotop`

this version, uses REST API, will be changed to use UDSF and websocket?

## required

* SensorBee: https://github.com/disktnk/sensorbee/tree/node-api
    * add `/node_status` API

## build & run

**single binary version**

```bash
$ cd $GOPATH/src/github.com/sensorbee/sensorbee-iotop
$ go build
$ ./sensorbee-iotop
```

**run as sensorbee command**

add "iotop" command to build.yaml like:

```yaml
commands:
  iotop:
    path: github.com/sensorbee/sensorbee-iotop/cmd
```

```bash
$ build_sensorbee --download-plugins=false
$ ./sensorbee iotop
```

## usage

**option**

- `d`: interval time [sec], default to 5 [sec]
- `uri`: URI address of target SensorBee server, default to "http://localhost:<default_port>/"
- `api-version`: version of SensorBee API, default to "v1"

**operation**

To stop, press "Ctrl+C" or "q"
