package main

import (
	"flag"
	"fmt"

	"github.com/stellentus/go-plc"
)

var addr = flag.String("address", "", "Hostname or IP of the PLC")
var path = flag.String("path", "1,0", "Path to the PLC at the provided host or IP")

func main() {
	connectionInfo := fmt.Sprintf("protocol=ab_eip&gateway=%s&path=%s&cpu=LGX", *addr, *path)
	timeout := 5000

	plc.New(connectionInfo, timeout)
}
