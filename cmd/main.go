package main

/*
#include <stdio.h>
#include <stdint.h>
*/
import "C"
import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"sync/atomic"
	"time"

	"github.com/stellentus/go-plc"
)

var addr = flag.String("address", "192.168.1.176", "Hostname or IP of the PLC")
var path = flag.String("path", "1,0", "Path to the PLC at the provided host or IP")
var workers = flag.Int("workers", 1, "Concurrent workers")
var tagfile = flag.String("tagfile", "", "File containing tags to peek and poke")
var timeout = flag.Int("timeout_ms", 5000, "Timeout (in ms)")
var iters = flag.Int("iters", 5000, "Number of IO loops for each thread to do")

var reads int64
var writes int64
var errors int64

func getListOfNames(allTags []plc.Tag) ([]string, error) {
	var ret []string

	if *tagfile == "" {
		// If no tagfile was provided, use the list of all tags
		for _, tag := range allTags {
			ret = append(ret, tag.Name())
		}
		return ret, nil
	}

	data, err := ioutil.ReadFile(*tagfile)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		ret = append(ret, line)
	}

	return ret, nil
}

func worker(ch chan bool) {
	connectionInfo := fmt.Sprintf("protocol=ab_eip&gateway=%s&path=%s&cpu=LGX", *addr, *path)

	thePlc, err := plc.New(connectionInfo, *timeout)
	if err != nil {
		panic(fmt.Sprintf("Can't set up the PLC: %q", err))
	}

	defer thePlc.Close()

	tags, err := thePlc.GetAllTags()
	if err != nil {
		panic(fmt.Sprintf("Can't get all tags: %q", err))
	}

	names, err := getListOfNames(tags)
	if err != nil {
		panic(fmt.Sprintf("Can't filter tags by tagfile: %q", err))
	}

	fmt.Printf("%q\n", names)

	n := *iters
	for i := 0; i < n; i++ {
		for _, name := range names {
			var val uint8

			err = thePlc.ReadTag(name, &val)
			atomic.AddInt64(&reads, 1)
			if err != nil {
				fmt.Printf("read to %s failed: %v\n", name, err)
				atomic.AddInt64(&errors, 1)
				continue
			}

			err = thePlc.WriteTag(name, val)
			atomic.AddInt64(&writes, 1)
			if err != nil {
				fmt.Printf("write to %s failed: %v\n", name, err)
				atomic.AddInt64(&errors, 1)
			}
		}
	}
	ch <- true
}
func main() {
	flag.Parse()

	ch := make(chan bool, *workers)
	b := time.Now()
	for i := 0; i < *workers; i++ {
		go worker(ch)
	}
	for i := 0; i < *workers; i++ {
		<-ch
	}
	e := time.Now()
	fmt.Printf("\nAll workers completed in %v.\n", e.Sub(b))
	fmt.Printf("Reads:  %v\n", atomic.LoadInt64(&reads))
	fmt.Printf("Writes: %v\n", atomic.LoadInt64(&writes))
	fmt.Printf("Errors: %v\n", atomic.LoadInt64(&errors))
}
