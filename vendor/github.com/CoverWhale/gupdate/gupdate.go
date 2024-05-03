package gupdate

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/minio/selfupdate"
)

type ReleaseGetter interface {
	getAllReleases() ([]Release, error)
	getLatestRelease() (Release, error)
}

type Release struct {
	Checksum string `json:"checksum,omitempty"`
	URL      string `json:"url"`
}

func GetAllReleases(r ReleaseGetter) ([]Release, error) {
	return r.getAllReleases()
}

func GetLatestRelease(r ReleaseGetter) (Release, error) {
	return r.getLatestRelease()
}

func (r Release) Update() error {
	resp, err := http.Get(r.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	cs, err := hex.DecodeString(r.Checksum)
	if err != nil {
		return err
	}

	if err := selfupdate.Apply(resp.Body, selfupdate.Options{
		Checksum: cs,
	}); err != nil {
		if updateErr := selfupdate.RollbackError(err); updateErr != nil {
			return fmt.Errorf("failed to rollback from bad update: %v", err)
		}

		return err
	}

	return nil
}
