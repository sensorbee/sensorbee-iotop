# Under implementing!

so interface and API will be changed

# sensorbee-iotop

## build & run

```bash
$ go get github.com/sensorbee/sensorbee-iotop
```

**single binary version**

```bash
$ cd $GOPATH/src/github.com/sensorbee/sensorbee-iotop
$ go build
$ ./sensorbee-iotop -t <topology_name>
```

**run as sensorbee command**

add "iotop" command to build.yaml like:

```yaml
commands:
  iotop:
    path: github.com/sensorbee/sensorbee-iotop/cmd
```

```bash
$ build_sensorbee
$ ./sensorbee iotop -t <topology_name>
```

## usage

**option**

- `-d`: interval time [sec], default to 5 [sec]
- `--uri`: URI address of target SensorBee server, default to "http://localhost:<default_port\>/"
- `--api-version`: version of SensorBee API, default to "v1"

**operation**

To stop, press "Ctrl+C" or "q"
