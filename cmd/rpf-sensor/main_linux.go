//go:build linux

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adilalperenciftci/agent-boundary/internal/rpf"
	"github.com/adilalperenciftci/agent-boundary/internal/sensor"
	"golang.org/x/sys/unix"
)

func main() {
	objectPath := flag.String("object", "", "compiled CO-RE BPF object")
	cgroupID := flag.Uint64("cgroup-id", 0, "target cgroup v2 ID")
	buildID := flag.String("build-id", "", "unique build execution ID")
	runID := flag.String("run-id", "", "CI or local run ID")
	bootID := flag.String("boot-id", "", "host boot ID")
	cgroupPathHash := flag.String("cgroup-path-hash", "", "SHA-256 commitment to registered cgroup path")
	outputPath := flag.String("output", "", "new canonical evidence JSONL file")
	flag.Parse()
	if *objectPath == "" || *cgroupID == 0 || *buildID == "" || *runID == "" ||
		*bootID == "" || *cgroupPathHash == "" || *outputPath == "" {
		fatal("--object, --output, --build-id, --run-id, --boot-id, --cgroup-path-hash, and non-zero --cgroup-id are required")
	}
	object, err := os.ReadFile(*objectPath)
	if err != nil {
		fatal("read BPF object: %v", err)
	}
	configMaterial := fmt.Sprintf("object_sha256=%s\ncgroup_id=%d\n", rpf.Digest(object), *cgroupID)
	source := rpf.Sensor{Name: "rpf-sensor", Version: "0.2.0", ConfigDigest: "sha256:" + rpf.Digest([]byte(configMaterial))}
	build := rpf.BuildScope{BuildID: *buildID, RunID: *runID, BootID: *bootID, CgroupID: *cgroupID, CgroupPathHash: *cgroupPathHash}
	chain, err := rpf.NewEventChain(build, source)
	if err != nil {
		fatal("initialize event chain: %v", err)
	}
	monitor, err := sensor.Open(*objectPath, *cgroupID)
	if err != nil {
		fatal("open sensor: %v", err)
	}
	defer monitor.Close()
	evidence, err := os.OpenFile(*outputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_APPEND, 0o600)
	if err != nil {
		fatal("create evidence stream: %v", err)
	}
	defer evidence.Close()
	writeEvent(evidence, chain, rpf.Event{
		ObservedAt: now(), MonotonicNS: boottimeNS(), Operation: "sensor_started",
		Resource: map[string]any{"target_cgroup_id": *cgroupID}, Outcome: rpf.Outcome{Status: "success"},
	})
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	for {
		event, err := monitor.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			fatal("read sensor: %v", err)
		}
		process := processFromExec(event, *bootID)
		writeEvent(evidence, chain, rpf.Event{
			ObservedAt: now(), MonotonicNS: event.MonotonicNS, Process: &process,
			Operation: "process_exec", Resource: map[string]any{}, Outcome: rpf.Outcome{Status: "success"},
		})
	}
	loss, err := monitor.Loss()
	counterReadError := err != nil
	writeEvent(evidence, chain, rpf.Event{
		ObservedAt: now(), MonotonicNS: boottimeNS(), Operation: "sensor_finalized",
		Resource: map[string]any{
			"kernel_reserve": loss, "decode": 0, "queue": 0, "persistence": 0,
			"counter_read_error": counterReadError,
		},
		Outcome: rpf.Outcome{Status: "success"},
	})
	if counterReadError {
		fatal("read loss counter: %v", err)
	}
}

func processFromExec(event sensor.ExecEvent, bootID string) rpf.Process {
	parentKey := rpf.ProcessKey(bootID, uint64(event.ParentPIDNamespace), event.PPID, event.ParentStartTimeNS)
	return rpf.Process{
		ProcessKey: rpf.ProcessKey(bootID, uint64(event.PIDNamespace), event.TGID, event.StartTimeNS),
		ParentKey:  parentKey, PID: event.PID, TGID: event.TGID, PPID: event.PPID,
		StartTimeNS: event.StartTimeNS, PIDNamespace: uint64(event.PIDNamespace),
		MountNamespace: uint64(event.MountNamespace), UID: event.UID, GID: event.GID,
		Executable: rpf.Executable{Path: event.Filename, IdentityKind: "path_only"},
	}
}

func writeEvent(output *os.File, chain *rpf.EventChain, event rpf.Event) {
	raw, err := chain.Append(event)
	if err != nil {
		fatal("chain event: %v", err)
	}
	if _, err := output.Write(raw); err != nil {
		fatal("persist event: %v", err)
	}
	if err := output.Sync(); err != nil {
		fatal("sync event: %v", err)
	}
}

func boottimeNS() uint64 {
	var value unix.Timespec
	if err := unix.ClockGettime(unix.CLOCK_BOOTTIME, &value); err != nil {
		fatal("read monotonic boot clock: %v", err)
	}
	return uint64(value.Sec)*1_000_000_000 + uint64(value.Nsec)
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func fatal(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", arguments...)
	os.Exit(4)
}
