## Overview

In an effort to better understand Apache Mesos, I built this Mesos framework to schedule a Docker container.  It takes a _very_ naive approach to scheduling, and simply initializes the container on the first Offer it comes in contact with.

## Requirements

* At least one Mesos Master and Agent
* [protoc](https://github.com/google/protobuf) if you want to modify the protbuf definitions

## Usage
```
$ mesos-fw
Minimal Mesos Framework

Usage:
  mesos-fw [command]

Available Commands:
  help        Help about any command
  launch      A brief description of your command

Flags:
  -h, --help   help for mesos-fw

Use "mesos-fw [command] --help" for more information about a command.
```