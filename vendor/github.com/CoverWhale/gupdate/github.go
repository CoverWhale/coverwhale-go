package gupdate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var githubReleaseURL = "https://api.github.com/repos"

type GitHubProject struct {
	Owner          string `json:"owner"`
	Name           string `json:"name,omitempty"`
	Platform       string `json:"platform"`
	Arch           string `json:"arch"`
	Checksum       string `json:"checksum"`
	CheckSumGetter CheckSumGetter
}

type GitHubReleases []GitHubRelease

type GitHubRelease struct {
	URL             string         `json:"url,omitempty"`
	AssetsURL       string         `json:"assets_url,omitempty"`
	UploadURL       string         `json:"upload_url,omitempty"`
	HTMLURL         string         `json:"html_url,omitempty"`
	ID              int            `json:"id,omitempty"`
	Author          GitHubAuthor   `json:"author,omitempty"`
	NodeID          string         `json:"node_id,omitempty"`
	TagName         string         `json:"tag_name,omitempty"`
	TargetCommitish string         `json:"target_commitish,omitempty"`
	Name            string         `json:"name,omitempty"`
	Draft           bool           `json:"draft,omitempty"`
	Prerelease      bool           `json:"prerelease,omitempty"`
	CreatedAt       time.Time      `json:"created_at,omitempty"`
	PublishedAt     time.Time      `json:"published_at,omitempty"`
	Assets          []GitHubAssets `json:"assets,omitempty"`
	TarballURL      string         `json:"tarball_url,omitempty"`
	ZipballURL      string         `json:"zipball_url,omitempty"`
	Body            string         `json:"body,omitempty"`
}
type GitHubAuthor struct {
	Login             string `json:"login,omitempty"`
	ID                int    `json:"id,omitempty"`
	NodeID            string `json:"node_id,omitempty"`
	AvatarURL         string `json:"avatar_url,omitempty"`
	GravatarID        string `json:"gravatar_id,omitempty"`
	URL               string `json:"url,omitempty"`
	HTMLURL           string `json:"html_url,omitempty"`
	FollowersURL      string `json:"followers_url,omitempty"`
	FollowingURL      string `json:"following_url,omitempty"`
	GistsURL          string `json:"gists_url,omitempty"`
	StarredURL        string `json:"starred_url,omitempty"`
	SubscriptionsURL  string `json:"subscriptions_url,omitempty"`
	OrganizationsURL  string `json:"organizations_url,omitempty"`
	ReposURL          string `json:"repos_url,omitempty"`
	EventsURL         string `json:"events_url,omitempty"`
	ReceivedEventsURL string `json:"received_events_url,omitempty"`
	Type              string `json:"type,omitempty"`
	SiteAdmin         bool   `json:"site_admin,omitempty"`
}
type GitHubUploader struct {
	Login             string `json:"login,omitempty"`
	ID                int    `json:"id,omitempty"`
	NodeID            string `json:"node_id,omitempty"`
	AvatarURL         string `json:"avatar_url,omitempty"`
	GravatarID        string `json:"gravatar_id,omitempty"`
	URL               string `json:"url,omitempty"`
	HTMLURL           string `json:"html_url,omitempty"`
	FollowersURL      string `json:"followers_url,omitempty"`
	FollowingURL      string `json:"following_url,omitempty"`
	GistsURL          string `json:"gists_url,omitempty"`
	StarredURL        string `json:"starred_url,omitempty"`
	SubscriptionsURL  string `json:"subscriptions_url,omitempty"`
	OrganizationsURL  string `json:"organizations_url,omitempty"`
	ReposURL          string `json:"repos_url,omitempty"`
	EventsURL         string `json:"events_url,omitempty"`
	ReceivedEventsURL string `json:"received_events_url,omitempty"`
	Type              string `json:"type,omitempty"`
	SiteAdmin         bool   `json:"site_admin,omitempty"`
}
type GitHubAssets struct {
	URL                string         `json:"url,omitempty"`
	ID                 int            `json:"id,omitempty"`
	NodeID             string         `json:"node_id,omitempty"`
	Name               string         `json:"name,omitempty"`
	Label              string         `json:"label,omitempty"`
	Uploader           GitHubUploader `json:"uploader,omitempty"`
	ContentType        string         `json:"content_type,omitempty"`
	State              string         `json:"state,omitempty"`
	Size               int            `json:"size,omitempty"`
	DownloadCount      int            `json:"download_count,omitempty"`
	CreatedAt          time.Time      `json:"created_at,omitempty"`
	UpdatedAt          time.Time      `json:"updated_at,omitempty"`
	BrowserDownloadURL string         `json:"browser_download_url,omitempty"`
}

func (g GitHubProject) getAllReleases() ([]Release, error) {
	var releases []Release
	var ghr []GitHubRelease
	url := fmt.Sprintf("%s/%s/%s/releases/latest", githubReleaseURL, g.Owner, g.Name)

	data, err := sendRequest(url)
	if err != nil {
		return releases, err
	}

	if err := json.Unmarshal(data, &ghr); err != nil {
		return releases, err
	}

	for _, release := range ghr {
		if len(release.Assets) == 0 {
			return releases, fmt.Errorf("no releases found")
		}
		for _, v := range release.Assets {
			if strings.Contains(v.Name, strings.ToLower(g.Platform)) && strings.Contains(v.Name, strings.ToLower(g.Arch)) {
				releases = append(releases, Release{URL: v.BrowserDownloadURL})
			}
		}
	}
	return nil, nil
}

func (g GitHubProject) getLatestRelease() (Release, error) {
	var release Release
	var ghr GitHubRelease
	url := fmt.Sprintf("%s/%s/%s/releases/latest", githubReleaseURL, g.Owner, g.Name)

	data, err := sendRequest(url)
	if err != nil {
		return release, err
	}

	if err := json.Unmarshal(data, &ghr); err != nil {
		return release, err
	}

	if len(ghr.Assets) == 0 {
		return release, fmt.Errorf("no releases found")
	}

	for _, v := range ghr.Assets {
		if strings.Contains(v.Name, strings.ToLower(g.Platform)) && strings.Contains(v.Name, strings.ToLower(g.Arch)) {
			release.URL = v.BrowserDownloadURL
		}
	}

	for _, v := range ghr.Assets {
		if strings.Contains(v.Name, "checksum") {
			checksum, err := g.getChecksums(v.BrowserDownloadURL)
			if err != nil {
				return release, err
			}

			release.Checksum = checksum
		}
	}

	if release.URL == "" {
		return release, fmt.Errorf("no results")
	}

	return release, nil
}

func (g GitHubProject) getChecksums(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return g.CheckSumGetter.GetChecksum(resp.Body)
}

func sendRequest(url string) ([]byte, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", `application/vnd.github+json`)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
