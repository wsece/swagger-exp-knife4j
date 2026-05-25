package scanner

import (
	"net/http"
	"strings"
)

// ProbeSummary aggregates automated GET/POST probe outcomes.
type ProbeSummary struct {
	Total        int
	OK           int // request sent and HTTP status received (no transport/build error)
	Skipped      int // non-GET/POST or explicitly skipped
	Failed       int // build error or transport failure
	Unauthorized int // StatusCode != 0 and != 401 (possible unauthenticated access)
}

// SummarizeProbeResults computes probe statistics from AutoRequestAllAPI results.
func SummarizeProbeResults(results []APIRequestResult) ProbeSummary {
	var s ProbeSummary
	s.Total = len(results)
	for _, r := range results {
		if isSkippedProbe(r) {
			s.Skipped++
		} else if r.Error != "" || r.StatusCode == 0 {
			s.Failed++
		} else {
			s.OK++
		}
		if r.StatusCode != 0 && r.StatusCode != http.StatusUnauthorized {
			s.Unauthorized++
		}
	}
	return s
}

func isSkippedProbe(r APIRequestResult) bool {
	return strings.HasPrefix(r.Error, "Skip")
}
