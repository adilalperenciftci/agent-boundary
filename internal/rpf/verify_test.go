package rpf

import (
	"bytes"
	"testing"
)

func fixtureInputs(t *testing.T, mutate func(*[]Event)) Inputs {
	t.Helper()
	artifact := []byte("deterministic fixture artifact\n")
	build := BuildScope{BuildID: "bld_fixture_001", RunID: "local-run-001", BootID: "4f25a5e2-3a0d-4bb0-99dd-a4e4b6c2a100", CgroupID: 99122, CgroupPathHash: "sha256:cgroup"}
	sensor := Sensor{Name: "rpf-fixture-sensor", Version: "0.1.0", ConfigDigest: "sha256:sensor-config"}
	parent := Process{ProcessKey: "sha256:parent", PID: 100, TGID: 100, PPID: 1, StartTimeNS: 1000, PIDNamespace: 42, MountNamespace: 43, UID: 1000, GID: 1000, Executable: Executable{Path: "/bin/sh", Identity: "sha256:sh", IdentityKind: "content_sha256"}}
	compiler := Process{ProcessKey: "sha256:compiler", ParentKey: parent.ProcessKey, PID: 101, TGID: 101, PPID: 100, StartTimeNS: 2000, PIDNamespace: 42, MountNamespace: 43, UID: 1000, GID: 1000, Executable: Executable{Path: "/usr/bin/cc", Identity: "sha256:cc", IdentityKind: "content_sha256"}}
	events := []Event{
		{SchemaVersion: EventSchema, EventID: "00000000-0000-7000-8000-000000000001", Sequence: 1, ObservedAt: "2026-09-11T12:00:00Z", MonotonicNS: 1, Build: build, Operation: "sensor_started", Outcome: Outcome{Status: "success"}, Sensor: sensor, Resource: map[string]any{}},
		{SchemaVersion: EventSchema, EventID: "00000000-0000-7000-8000-000000000002", Sequence: 2, ObservedAt: "2026-09-11T12:00:01Z", MonotonicNS: 2, Build: build, Process: &parent, Operation: "process_exec", Outcome: Outcome{Status: "success"}, Sensor: sensor, Resource: map[string]any{}},
		{SchemaVersion: EventSchema, EventID: "00000000-0000-7000-8000-000000000003", Sequence: 3, ObservedAt: "2026-09-11T12:00:02Z", MonotonicNS: 3, Build: build, Process: &compiler, Operation: "process_exec", Outcome: Outcome{Status: "success"}, Sensor: sensor, Resource: map[string]any{}},
		{SchemaVersion: EventSchema, EventID: "00000000-0000-7000-8000-000000000004", Sequence: 4, ObservedAt: "2026-09-11T12:00:03Z", MonotonicNS: 4, Build: build, Process: &compiler, Operation: "artifact_finalized", Outcome: Outcome{Status: "success"}, Sensor: sensor, Resource: map[string]any{"category": "artifact", "sha256": Digest(artifact)}},
		{SchemaVersion: EventSchema, EventID: "00000000-0000-7000-8000-000000000005", Sequence: 5, ObservedAt: "2026-09-11T12:00:04Z", MonotonicNS: 5, Build: build, Operation: "sensor_finalized", Outcome: Outcome{Status: "success"}, Sensor: sensor, Resource: map[string]any{"kernel_reserve": 0, "kernel_correlation": 0, "decode": 0, "queue": 0, "persistence": 0, "counter_read_error": false}},
	}
	if mutate != nil {
		mutate(&events)
	}
	previous := zeroHash
	var stream bytes.Buffer
	for index := range events {
		events[index].Sequence = uint64(index + 1)
		events[index].Integrity.PreviousEventHash = previous
		hash, err := eventDigest(events[index])
		if err != nil {
			t.Fatal(err)
		}
		events[index].Integrity.EventHash = hash
		previous = hash
		raw, err := canonical(events[index])
		if err != nil {
			t.Fatal(err)
		}
		stream.Write(raw)
		stream.WriteByte('\n')
	}
	identity := Correlation{BuildID: build.BuildID, RunIdentity: RunIdentity{Provider: "local", RunID: build.RunID, Attempt: 1}, Source: SourceIdentity{Repository: "https://example.test/repo", Revision: "1111111111111111111111111111111111111111"}}
	identityMap, err := toMap(identity)
	if err != nil {
		t.Fatal(err)
	}
	provenance := Statement{Type: StatementType, Subject: []Subject{{Name: "artifact", Digest: map[string]string{"sha256": Digest(artifact)}}}, PredicateType: SLSAPredicate, Predicate: map[string]any{"buildDefinition": map[string]any{"internalParameters": map[string]any{BuildIdentityKey: identityMap}}}}
	provenanceBytes, err := canonical(provenance)
	if err != nil {
		t.Fatal(err)
	}
	policy := Policy{SchemaVersion: "0.1", AllowedExecutables: []string{"/bin/sh", "/usr/bin/cc"}, AllowedNetworkDestinations: []string{"127.0.0.1:8080"}, ForbiddenSensitiveCategories: []string{"synthetic_credential"}, ExpectedProvider: "local", IncompleteDecision: "REJECT"}
	policyBytes, err := canonical(policy)
	if err != nil {
		t.Fatal(err)
	}
	return Inputs{ArtifactBytes: artifact, ArtifactName: "artifact", EventBytes: stream.Bytes(), ProvenanceBytes: provenanceBytes, PolicyBytes: policyBytes}
}

func TestBaselineAssemblesAndAllows(t *testing.T) {
	inputs := fixtureInputs(t, nil)
	bundle, err := Assemble(inputs)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := Verify(inputs, bundle)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Decision != "ALLOW" || decision.Completeness != "complete" {
		t.Fatalf("unexpected decision: %#v", decision)
	}
	if len(bundle.Graph.Nodes) != 2 || len(bundle.Graph.Edges) != 2 {
		t.Fatalf("unexpected graph: %#v", bundle.Graph)
	}
}

func TestBundleDiskRoundTrip(t *testing.T) {
	inputs := fixtureInputs(t, nil)
	bundle, err := Assemble(inputs)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if err := WriteBundle(directory, bundle); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadBundle(directory)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := Verify(inputs, loaded)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Decision != "ALLOW" {
		t.Fatalf("disk bundle failed verification: %#v", decision)
	}
}

func TestEventLossCannotAllow(t *testing.T) {
	inputs := fixtureInputs(t, func(events *[]Event) { (*events)[len(*events)-1].Resource["kernel_reserve"] = 1 })
	bundle, err := Assemble(inputs)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := Verify(inputs, bundle)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Decision != "REJECT" || decision.Completeness != "incomplete" {
		t.Fatalf("loss was accepted: %#v", decision)
	}
}

func TestSensitiveAccessAndUnexpectedEgressReject(t *testing.T) {
	for _, operation := range []string{"file_open_sensitive", "network_connect"} {
		caseInputs := fixtureInputs(t, func(events *[]Event) {
			base := *events
			behavior := base[3]
			behavior.EventID = "00000000-0000-7000-8000-000000000006"
			behavior.Operation = operation
			if operation == "file_open_sensitive" {
				behavior.Resource = map[string]any{"category": "synthetic_credential"}
			} else {
				behavior.Resource = map[string]any{"destination": "127.0.0.1:9090"}
			}
			*events = append(base[:4], append([]Event{behavior}, base[4:]...)...)
		})
		bundle, err := Assemble(caseInputs)
		if err != nil {
			t.Fatal(err)
		}
		decision, err := Verify(caseInputs, bundle)
		if err != nil {
			t.Fatal(err)
		}
		if decision.Decision != "REJECT" {
			t.Fatalf("%s was accepted: %#v", operation, decision)
		}
	}
}

func TestArtifactSubstitutionFails(t *testing.T) {
	inputs := fixtureInputs(t, nil)
	bundle, err := Assemble(inputs)
	if err != nil {
		t.Fatal(err)
	}
	inputs.ArtifactBytes = []byte("substituted\n")
	decision, err := Verify(inputs, bundle)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Decision != "REJECT" {
		t.Fatalf("substitution was accepted: %#v", decision)
	}
}

func TestManifestTamperFails(t *testing.T) {
	inputs := fixtureInputs(t, nil)
	bundle, err := Assemble(inputs)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Manifest.BuildID = "other-build"
	decision, err := Verify(inputs, bundle)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Decision != "REJECT" {
		t.Fatalf("tamper was accepted: %#v", decision)
	}
}

func TestDuplicateJSONKeyRejected(t *testing.T) {
	var policy Policy
	err := decodeStrict([]byte(`{"schema_version":"0.1","schema_version":"0.1"}`), &policy, maxDocumentBytes)
	if err == nil {
		t.Fatal("duplicate key accepted")
	}
}
