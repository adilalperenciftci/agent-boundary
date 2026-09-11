//go:build linux

package main

import "testing"

func TestIPv4Destination(t *testing.T) {
	if got := ipv4Destination(0x0100007f, 18080); got != "127.0.0.1:18080" {
		t.Fatalf("got %q", got)
	}
}
