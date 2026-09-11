//go:build linux

package sensor

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
)

const (
	commLength = 16
	pathLength = 256
)

type ExecEvent struct {
	MonotonicNS uint64 `json:"monotonic_ns"`
	CgroupID    uint64 `json:"cgroup_id"`
	PID         uint32 `json:"pid"`
	TGID        uint32 `json:"tgid"`
	PPID        uint32 `json:"ppid"`
	UID         uint32 `json:"uid"`
	GID         uint32 `json:"gid"`
	Command     string `json:"command"`
	Filename    string `json:"filename"`
}

type wireExecEvent struct {
	MonotonicNS uint64
	CgroupID    uint64
	PID         uint32
	TGID        uint32
	PPID        uint32
	UID         uint32
	GID         uint32
	Command     [commLength]byte
	Filename    [pathLength]byte
	Padding     [4]byte
}

type Sensor struct {
	collection *ebpf.Collection
	tracepoint link.Link
	reader     *ringbuf.Reader
}

func Open(objectPath string, cgroupID uint64) (*Sensor, error) {
	if cgroupID == 0 {
		return nil, errors.New("target cgroup ID must be non-zero")
	}
	spec, err := ebpf.LoadCollectionSpec(objectPath)
	if err != nil {
		return nil, fmt.Errorf("load BPF object: %w", err)
	}
	target, ok := spec.Variables["target_cgroup_id"]
	if !ok {
		return nil, errors.New("BPF object lacks target_cgroup_id variable")
	}
	if err := target.Set(cgroupID); err != nil {
		return nil, fmt.Errorf("set target cgroup: %w", err)
	}
	collection, err := ebpf.NewCollection(spec)
	if err != nil {
		return nil, fmt.Errorf("load BPF collection: %w", err)
	}
	program := collection.Programs["observe_exec"]
	if program == nil {
		collection.Close()
		return nil, errors.New("BPF object lacks observe_exec program")
	}
	attached, err := link.Tracepoint("sched", "sched_process_exec", program, nil)
	if err != nil {
		collection.Close()
		return nil, fmt.Errorf("attach exec tracepoint: %w", err)
	}
	events := collection.Maps["events"]
	if events == nil {
		attached.Close()
		collection.Close()
		return nil, errors.New("BPF object lacks events map")
	}
	reader, err := ringbuf.NewReader(events)
	if err != nil {
		attached.Close()
		collection.Close()
		return nil, fmt.Errorf("open events ring buffer: %w", err)
	}
	return &Sensor{collection: collection, tracepoint: attached, reader: reader}, nil
}

func (sensor *Sensor) Read(ctx context.Context) (ExecEvent, error) {
	if sensor == nil || sensor.reader == nil {
		return ExecEvent{}, errors.New("sensor is not open")
	}
	for {
		if err := ctx.Err(); err != nil {
			return ExecEvent{}, err
		}
		deadline := time.Now().Add(250 * time.Millisecond)
		if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(deadline) {
			deadline = contextDeadline
		}
		sensor.reader.SetDeadline(deadline)
		record, err := sensor.reader.Read()
		if errors.Is(err, os.ErrDeadlineExceeded) {
			continue
		}
		if err != nil {
			return ExecEvent{}, err
		}
		return decodeExec(record.RawSample)
	}
}

func decodeExec(raw []byte) (ExecEvent, error) {
	var wire wireExecEvent
	if len(raw) != binary.Size(wire) {
		return ExecEvent{}, fmt.Errorf("exec sample has size %d, expected %d", len(raw), binary.Size(wire))
	}
	if err := binary.Read(bytes.NewReader(raw), binary.LittleEndian, &wire); err != nil {
		return ExecEvent{}, fmt.Errorf("decode exec sample: %w", err)
	}
	return ExecEvent{
		MonotonicNS: wire.MonotonicNS, CgroupID: wire.CgroupID,
		PID: wire.PID, TGID: wire.TGID, PPID: wire.PPID, UID: wire.UID, GID: wire.GID,
		Command: cString(wire.Command[:]), Filename: cString(wire.Filename[:]),
	}, nil
}

func cString(value []byte) string {
	if index := bytes.IndexByte(value, 0); index >= 0 {
		value = value[:index]
	}
	return string(value)
}

func (sensor *Sensor) Loss() (uint64, error) {
	if sensor == nil || sensor.collection == nil {
		return 0, errors.New("sensor is not open")
	}
	lossMap := sensor.collection.Maps["ringbuf_drops"]
	if lossMap == nil {
		return 0, errors.New("BPF object lacks loss map")
	}
	var perCPU []uint64
	if err := lossMap.Lookup(uint32(0), &perCPU); err != nil {
		return 0, fmt.Errorf("read ring-buffer loss: %w", err)
	}
	var total uint64
	for _, value := range perCPU {
		total += value
	}
	return total, nil
}

func (sensor *Sensor) Close() error {
	if sensor == nil {
		return nil
	}
	var result error
	if sensor.reader != nil {
		result = errors.Join(result, sensor.reader.Close())
	}
	if sensor.tracepoint != nil {
		result = errors.Join(result, sensor.tracepoint.Close())
	}
	if sensor.collection != nil {
		sensor.collection.Close()
	}
	return result
}

var _ io.Closer = (*Sensor)(nil)
