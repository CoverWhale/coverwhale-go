// Package gosecdemo is a deliberate-finding fixture for the DOF-139 gosec CI gate.
// This package is intentionally NOT used anywhere in coverwhale-go runtime code.
// It exists solely to demonstrate that golangci-lint + gosec correctly fail CI on
// security-relevant findings. The accompanying PR is intended to be closed without
// merging — see the PR description.
package gosecdemo

import "fmt"

// hardcoded credential — gosec G101 should flag this
const apiKey = "AKIA1234567890ABCDEF1234"

// Run uses the hardcoded credential. This function is unreachable in production.
func Run() {
	fmt.Println("api key:", apiKey)
}
