package rpf

import (
	"bytes"
	"testing"
)

func TestEventChainProducesParseableCanonicalStream(t *testing.T) {
	chain, err := NewEventChain(
		BuildScope{BuildID: "build-1", RunID: "run-1", BootID: "boot-1", CgroupID: 7, CgroupPathHash: "sha256:path"},
		Sensor{Name: "sensor", Version: "0.1.0", ConfigDigest: "sha256:config"},
	)
	if err != nil {
		t.Fatal(err)
	}
	events := []Event{
		{ObservedAt: "2026-09-11T12:00:00Z", MonotonicNS: 10, Operation: "sensor_started", Resource: map[string]any{}, Outcome: Outcome{Status: "success"}},
		{ObservedAt: "2026-09-11T12:00:01Z", MonotonicNS: 11, Operation: "sensor_finalized", Resource: map[string]any{"kernel_reserve": 0, "decode": 0, "queue": 0, "persistence": 0, "counter_read_error": false}, Outcome: Outcome{Status: "success"}},
	}
	var stream bytes.Buffer
	for _, event := range events {
		raw, appendErr := chain.Append(event)
		if appendErr != nil {
			t.Fatal(appendErr)
		}
		stream.Write(raw)
	}
	parsed, err := ParseEventStream(stream.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 2 || parsed[0].Sequence != 1 || parsed[1].Sequence != 2 ||
		parsed[1].Integrity.PreviousEventHash != parsed[0].Integrity.EventHash {
		t.Fatalf("unexpected parsed chain: %+v", parsed)
	}
}

func TestEventChainRejectsCallerOwnedMetadata(t *testing.T) {
	chain, err := NewEventChain(
		BuildScope{BuildID: "build-1", RunID: "run-1", BootID: "boot-1", CgroupID: 7, CgroupPathHash: "sha256:path"},
		Sensor{Name: "sensor", Version: "0.1.0", ConfigDigest: "sha256:config"},
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = chain.Append(Event{Sequence: 9})
	if err == nil {
		t.Fatal("expected caller-owned metadata rejection")
	}
}

func TestProcessKeyRequiresCompositeIdentity(t *testing.T) {
	if ProcessKey("boot", 3, 4, 5) == "" {
		t.Fatal("complete identity did not produce a key")
	}
	if ProcessKey("boot", 0, 4, 5) != "" {
		t.Fatal("incomplete identity produced a key")
	}
}
