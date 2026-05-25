package version

import "fmt"

// Report returns version metadata lines for CLI `version` output.
func Report() []string {
	return []string{
		"swagger-exp-knife4j",
		fmt.Sprintf("Version:    %s", Version),
		fmt.Sprintf("Git commit: %s", GitHash),
		fmt.Sprintf("Built:      %s", GoBuildTime),
		fmt.Sprintf("Build env:  %s", GoBuildEnv),
	}
}
