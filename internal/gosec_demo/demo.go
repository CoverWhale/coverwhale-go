// Package gosecdemo is a deliberate-finding fixture for the DOF-139 gosec CI gate.
// This package is intentionally NOT used anywhere in coverwhale-go runtime code.
// It exists solely to demonstrate that golangci-lint + gosec correctly fail CI on
// security-relevant findings.
package gosecdemo

import (
	"crypto/md5"  //nolint:gci // gosec G501 demo
	"database/sql"
	"fmt"
	"math/rand"
)

// WeakHash uses MD5 — gosec G501 (blocklisted import) and G401 (weak crypto)
func WeakHash(data []byte) [16]byte {
	return md5.Sum(data)
}

// WeakRandom uses math/rand for token-like ID — gosec G404 (weak random source)
func WeakRandom() int {
	return rand.Int()
}

// SqlInjection — gosec G201 (SQL string formatting)
func SqlInjection(db *sql.DB, userInput string) (*sql.Rows, error) {
	query := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", userInput)
	return db.Query(query)
}
