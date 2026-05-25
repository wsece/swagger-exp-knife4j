package cmd

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
)

func formatReportListenError(host string, port int, err error) error {
	if err == nil {
		return nil
	}
	if isTCPAddrInUse(err) {
		return fmt.Errorf(
			"port %d on %s is already in use (another report server or process may be holding it)\n"+
				"  try another port: swagger-exp-knife4j report server --port %d\n"+
				"  or stop the process using port %d, then retry\n\n(%v)",
			port, host, port+1, port, err,
		)
	}
	return fmt.Errorf("listen on %s:%d: %w", host, port, err)
}

func isTCPAddrInUse(err error) bool {
	if errors.Is(err, syscall.EADDRINUSE) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "address already in use") ||
		strings.Contains(msg, "only one usage of each socket address")
}
