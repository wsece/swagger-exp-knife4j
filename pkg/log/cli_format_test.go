package log

import (
	"fmt"
	"testing"
)

func TestScanErrorMessage_resolveSwagger(t *testing.T) {
	msg := ScanErrorMessage(
		fmt.Errorf("resolve swagger url: %w", fmt.Errorf("[-] Can't found the Swagger JSON URL. Please check the input URL or try to specify the JSON URL directly.")),
	)
	want := "Can't found the Swagger JSON URL. Please check the input URL or try to specify the JSON URL directly."
	if msg != want {
		t.Fatalf("got %q want %q", msg, want)
	}
}
