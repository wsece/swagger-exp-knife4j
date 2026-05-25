package cmd

import (
	"errors"
	"strings"
	"syscall"
	"testing"
)

func TestIsTCPAddrInUse(t *testing.T) {
	if !isTCPAddrInUse(syscall.EADDRINUSE) {
		t.Fatal("EADDRINUSE")
	}
	if !isTCPAddrInUse(errors.New(`listen tcp 127.0.0.1:7171: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted.`)) {
		t.Fatal("windows message")
	}
	if isTCPAddrInUse(errors.New("connection refused")) {
		t.Fatal("unexpected")
	}
}

func TestFormatReportListenError_portInUse(t *testing.T) {
	err := formatReportListenError("127.0.0.1", 7171, syscall.EADDRINUSE)
	if err == nil {
		t.Fatal("expected error")
	}
	s := err.Error()
	if !strings.Contains(s, "--port 7172") {
		t.Fatalf("missing port hint: %s", s)
	}
}
