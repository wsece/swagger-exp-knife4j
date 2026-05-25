package scanner

import "testing"

func TestSummarizeProbeResults_skipFailOk(t *testing.T) {
	results := []APIRequestResult{
		{StatusCode: 200},
		{Error: "Skip non-GET/POST methods"},
		{Error: "connection refused", StatusCode: 0},
	}
	s := SummarizeProbeResults(results)
	if s.Total != 3 || s.OK != 1 || s.Skipped != 1 || s.Failed != 1 {
		t.Fatalf("got %+v", s)
	}
}
