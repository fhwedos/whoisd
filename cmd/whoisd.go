// Copyright 2017 Openprovider Authors. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/fhwedos/whoisd/pkg/config"
	"github.com/fhwedos/whoisd/pkg/memcache"
	"github.com/fhwedos/whoisd/pkg/service"
	"github.com/fhwedos/whoisd/pkg/version"
)

var (
	stdlog, errlog *log.Logger
)

// Init "Usage" helper
func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	stdlog = log.New(os.Stdout, "[SETUP]: ", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "[SETUP:ERROR]: ", log.Ldate|log.Ltime|log.Lshortfile)
	flag.Usage = func() {
		fmt.Println(config.Usage())
	}
}

func main() {
	daemonName, daemonDescription := "whoisd", "Whois Daemon"
	daemon, err := service.New(daemonName, daemonDescription)
	if err != nil {
		errlog.Println("Error: ", err)
		os.Exit(1)
	}
	flag.Parse()
	if daemon.Config.ShowVersion {
		buildTime, err := time.Parse(time.RFC3339, version.DATE)
		if err != nil {
			buildTime = time.Now()
		}
		fmt.Println(daemonName, version.RELEASE, buildTime.Format(time.RFC1123))
		os.Exit(0)
	}

	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "cacheFlush":
			err := memcache.WriteCacheControl(daemon.Config, true, false)
			if err != nil {
				fmt.Println("ERROR: Flushing cache failed.", err)
			} else {
				fmt.Println("Cache will be flushed.")
			}
			os.Exit(0)
		case "cacheStatus":
			err := memcache.WriteCacheControl(daemon.Config, false, true)
			if err != nil {
				fmt.Println("ERROR: Failed to list cache items.")
			} else {
				fmt.Println("Cache items will be listed in file: ", daemon.Config.CacheListFile)
			}
			os.Exit(0)
		}
	}

	status, err := daemon.Run()
	if err != nil {
		errlog.Printf("%s - %s", status, err)
		os.Exit(1)
	}
	fmt.Println(status)
}
