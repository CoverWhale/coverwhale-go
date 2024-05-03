package gupdate

import (
	"bufio"
	"fmt"
	"io"
	"runtime"
	"strings"
)

type CheckSumGetter interface {
	GetChecksum(io.Reader) (string, error)
}

type Goreleaser struct{}

func (g Goreleaser) GetChecksum(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, strings.ToLower(runtime.GOOS)) && strings.Contains(line, strings.ToLower(runtime.GOARCH)) {
			return strings.Split(line, " ")[0], nil
		}
	}

	return "", fmt.Errorf("valid checksum not found")
}
