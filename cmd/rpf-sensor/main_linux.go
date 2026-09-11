//go:build linux

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/adilalperenciftci/agent-boundary/internal/sensor"
)

func main() {
	objectPath := flag.String("object", "", "compiled CO-RE BPF object")
	cgroupID := flag.Uint64("cgroup-id", 0, "target cgroup v2 ID")
	flag.Parse()
	if *objectPath == "" || *cgroupID == 0 {
		fatal("--object and non-zero --cgroup-id are required")
	}
	monitor, err := sensor.Open(*objectPath, *cgroupID)
	if err != nil {
		fatal("open sensor: %v", err)
	}
	defer monitor.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	encoder := json.NewEncoder(os.Stdout)
	for {
		event, err := monitor.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			fatal("read sensor: %v", err)
		}
		if err := encoder.Encode(event); err != nil {
			fatal("encode event: %v", err)
		}
	}
	loss, err := monitor.Loss()
	if err != nil {
		fatal("read loss counter: %v", err)
	}
	if err := encoder.Encode(map[string]any{"sensor_finalized": true, "ringbuf_drops": loss}); err != nil {
		fatal("encode final state: %v", err)
	}
}

func fatal(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", arguments...)
	os.Exit(4)
}
