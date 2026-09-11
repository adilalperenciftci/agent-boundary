//go:build linux

package sensor

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
)

func TestDecodeExec(t *testing.T) {
	wire := wireExecEvent{
		MonotonicNS:        42,
		CgroupID:           73,
		StartTimeNS:        40,
		ParentStartTimeNS:  30,
		PID:                101,
		TGID:               100,
		PPID:               12,
		UID:                1000,
		GID:                1001,
		PIDNamespace:       7,
		MountNamespace:     8,
		ParentPIDNamespace: 6,
	}
	copy(wire.Command[:], "compiler")
	copy(wire.Filename[:], "/usr/bin/cc")
	var encoded bytes.Buffer
	if err := binary.Write(&encoded, binary.LittleEndian, wire); err != nil {
		t.Fatal(err)
	}

	event, err := decodeExec(encoded.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if event.MonotonicNS != 42 || event.CgroupID != 73 || event.StartTimeNS != 40 ||
		event.ParentStartTimeNS != 30 || event.PID != 101 ||
		event.TGID != 100 || event.PPID != 12 || event.UID != 1000 || event.GID != 1001 ||
		event.PIDNamespace != 7 || event.MountNamespace != 8 || event.ParentPIDNamespace != 6 ||
		event.Command != "compiler" || event.Filename != "/usr/bin/cc" {
		t.Fatalf("unexpected decoded event: %+v", event)
	}
}

func TestDecodeExecRejectsWrongSize(t *testing.T) {
	_, err := decodeExec(make([]byte, binary.Size(wireExecEvent{})-1))
	if err == nil || !strings.Contains(err.Error(), "expected") {
		t.Fatalf("expected bounded-size error, got %v", err)
	}
}

func TestCStringWithoutTerminatorUsesBoundedInput(t *testing.T) {
	if got := cString([]byte("1234")); got != "1234" {
		t.Fatalf("got %q", got)
	}
}
