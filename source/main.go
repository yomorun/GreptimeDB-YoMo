package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/yomorun/yomo"
)

var (
	tag        = uint32(0xC001)
	credential = flag.String("c", "", "yomo zipper credential")
	zipperAddr = flag.String("z", "127.0.0.1:9000", "yomo zipper address")
)

func main() {
	flag.Parse()

	source, err := newSource(*zipperAddr, *credential)
	if err != nil {
		log.Fatalln(err)
	}

	// generate logs every 1 second, the format is `monitor,host=Darwin user_cpu=3.35,sys_cpu=9.71,idle_cpu=86.63,memory=3056 1721541370000000000`
	// Define the log format
	logFormat := "monitor,host=vm-uswest-3 user_cpu=%f,sys_cpu=%f,idle_cpu=%f %d"

	// Create a ticker that fires every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	before, _ := cpu.Get()

	for range ticker.C {
		after, _ := cpu.Get()
		user := float64(after.User-before.User) / float64(after.Total-before.Total) * 100
		sys := float64(after.System-before.System) / float64(after.Total-before.Total) * 100
		idle := float64(after.Idle-before.Idle) / float64(after.Total-before.Total) * 100
		// Generate and print a log entry
		data := fmt.Sprintf(logFormat, user, sys, idle, time.Now().UnixNano())
		log.Println(data)
		if err := source.Write(tag, []byte(data)); err != nil {
			log.Println(err)
		}
		before = after
	}
}

func newSource(zipperAddr string, credential string) (yomo.Source, error) {
	opts := []yomo.SourceOption{
		yomo.WithSourceReConnect(),
	}
	if credential != "" {
		opts = append(opts, yomo.WithCredential(credential))
	}
	source := yomo.NewSource(
		"log-generator",
		zipperAddr,
		opts...,
	)
	if err := source.Connect(); err != nil {
		return nil, source.Connect()
	}
	return source, nil
}
