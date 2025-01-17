/*
 * Copyright 2019 The NATS Authors
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/nats-io/nats-kafka/server/core"
)

func main() {
	var server *core.NATSKafkaBridge
	var err error

	flags := core.Flags{}
	flag.StringVar(&flags.ConfigFile, "c", "", "configuration filepath, other flags take precedent over the config file, can be set with $NATS_KAFKA_BRIDGE_CONFIG")
	flag.BoolVar(&flags.Debug, "D", false, "turn on debug logging")
	flag.BoolVar(&flags.Verbose, "V", false, "turn on verbose logging")
	flag.BoolVar(&flags.DebugAndVerbose, "DV", false, "turn on debug and verbose logging")
	flag.Parse()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGHUP)

		for {
			signal := <-sigChan

			if signal == os.Interrupt {
				if server.Logger() != nil {
					fmt.Println() // clear the line for the control-C
					server.Logger().Noticef("received sig-interrupt, shutting down")
				}
				//server.Stop()
				os.Exit(0)
			}

			if signal == syscall.SIGHUP {
				if server.Logger() != nil {
					server.Logger().Errorf("received sig-hup, restarting")
				}
				server.Stop()
				server := core.NewNATSKafkaBridge()

				err = server.InitializeFromFlags(flags)
				if err != nil {
					logErrorFatal(server, "error initializing config, %s", err.Error())
				}

				err = server.Start()
				if err != nil {
					logErrorFatal(server, "error starting bridge, %s", err.Error())
				}
			}
		}
	}()

	server = core.NewNATSKafkaBridge()
	err = server.InitializeFromFlags(flags)
	if err != nil {
		logErrorFatal(server, "error initializing config, %s", err.Error())
	}

	err = server.Start()
	if err != nil {
		logErrorFatal(server, "error starting bridge, %s", err.Error())
	}

	// exit main but keep running goroutines
	runtime.Goexit()
}

func logErrorFatal(b *core.NATSKafkaBridge, format string, args ...interface{}) {
	if b.Logger() != nil {
		b.Logger().Errorf(format, args...)
	} else {
		log.Printf(format, args...)
	}
	b.Stop()
	os.Exit(0)
}
