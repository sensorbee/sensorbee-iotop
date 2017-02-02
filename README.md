Current version is not stable, and it is possible to change view layout, command arguments and so on.

# sensorbee-iotop

## build & run

```bash
$ go get github.com/sensorbee/sensorbee-iotop
```

### single binary version

```bash
$ cd $GOPATH/src/github.com/sensorbee/sensorbee-iotop
$ go build
$ ./sensorbee-iotop -t <topology_name>
```

### run as sensorbee command

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

### command option

- `-d`: interval time [sec], default to 5 [sec]
- `-c`: view total count on in/out, default to `false` and show by [tuples/sec]
- `-u`: select node type to show, input node type name, default to "" means "all"
- `--uri`: URI address of target SensorBee server, default to `http://localhost:<default_port>`
- `--api-version`: version of SensorBee API, default to "v1"

### operation (on running)

- `d`: change interval time
- `c`: change in/out unit, which "total count of tuples" or "[tupels/sec]"
- `u`: change which node type to show
- `q` or `Ctrl+C`: stop iotop process
