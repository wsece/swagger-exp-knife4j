// http_options_test.go: tests for HTTPOptions.Validate and Client timeout behavior.
package scanner

import (
	"testing"
	"time"
)

func TestHTTPOptions_Validate(t *testing.T) {
	t.Parallel()
	o := &HTTPOptions{Parallel: 1, ConnectTimeout: 30 * time.Second}
	if err := o.Validate(); err != nil {
		t.Fatal(err)
	}
	o.Parallel = 0
	if err := o.Validate(); err == nil {
		t.Fatal("expected error for parallel=0")
	}
}

func TestHTTPOptions_Client_requestTimeoutOverride(t *testing.T) {
	t.Parallel()
	o := &HTTPOptions{
		RequestTimeout: 5 * time.Second,
		ConnectTimeout: 10 * time.Second,
	}
	c, err := o.Client(15 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if c.Timeout != 5*time.Second {
		t.Fatalf("timeout %v want 5s", c.Timeout)
	}
}
