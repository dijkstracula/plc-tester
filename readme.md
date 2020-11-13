# PLC-Tester

## Setup

Initial setup for PLC-Tester is slightly convoluted in that a dependency,
`go-plc`, requires an object file to be dropped in its source path

```
$ export GO111MODULE=on
$ export GOPROXY=direct
$ export GOSUMDB=off

To pull private Git repositories via ssh rather than over http:
$ git config --global --add url."git@github.com:".insteadOf "https://github.com/"
```

```
$ go get -u ./...
$ cp ../path/to/libplctag.a vendor/github.com/stellentus/go-plc/
