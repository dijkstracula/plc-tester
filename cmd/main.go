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

	"github.com/stellentus/go-plc"
)

var addr = flag.String("address", "192.168.29.121", "Hostname or IP of the PLC")
var path = flag.String("path", "1,0", "Path to the PLC at the provided host or IP")
var workers = flag.Int("workers", 1, "Concurrent workers")
var tagfile = flag.String("tagfile", "", "File containing tags to peek and poke")

var reads int64
var writes int64
var errors int64

func filterTagsByTagfile(allTags []plc.Tag) ([]plc.Tag, error) {
	var byName = make(map[string]plc.Tag)
	for _, tag := range allTags {
		byName[tag.Name()] = tag
	}
	var ret []plc.Tag

	if *tagfile == "" {
		return allTags, nil
	}

	data, err := ioutil.ReadFile(*tagfile)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		tag, ok := byName[line]
		if !ok {
			return nil, fmt.Errorf("Tag %s missing on the PLC", line)
		}
		ret = append(ret, tag)
	}

	return ret, nil
}

func worker(ch chan bool) {
	connectionInfo := fmt.Sprintf("protocol=ab_eip&gateway=%s&path=%s&cpu=LGX", *addr, *path)
	timeout := 5000

	thePlc, err := plc.New(connectionInfo, timeout)
	if err != nil {
		panic(fmt.Sprintf("Can't set up the PLC: %q", err))
	}

	defer thePlc.Close()

	tags, err := thePlc.GetAllTags()
	if err != nil {
		panic(fmt.Sprintf("Can't get all tags: %q", err))
	}

	tags, err = filterTagsByTagfile(tags)
	if err != nil {
		panic(fmt.Sprintf("Can't filter tags by tagfile: %q", err))
	}

	fmt.Printf("%q\n", tags)

	for _, tag := range tags {
		var val uint8

		err = thePlc.ReadTag(tag.Name(), val)
		atomic.AddInt64(&reads, 1)
		if err != nil {
			atomic.AddInt64(&errors, 1)
		}

		fmt.Printf("%s: %v\n", tag.Name(), val)
	}
	ch <- true
}

func main() {
	flag.Parse()

	ch := make(chan bool, *workers)
	for i := 0; i < *workers; i++ {
		go worker(ch)
	}
	for i := 0; i < *workers; i++ {
		<-ch
	}
	fmt.Printf("All workers completed.\n")
	fmt.Printf("Reads:  %ld\n", atomic.LoadInt64(&reads))
	fmt.Printf("Writes: %ld\n", atomic.LoadInt64(&writes))
	fmt.Printf("Errors: %ld\n", atomic.LoadInt64(&errors))
}
