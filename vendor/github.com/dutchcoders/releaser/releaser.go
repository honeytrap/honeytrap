package releaser

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-semver/semver"
)

type Releaser struct {
	Client *http.Client

	owner string
	repo  string
}

func New(owner, repo string) *Releaser {
	return &Releaser{
		Client: http.DefaultClient,
		owner:  owner,
		repo:   repo,
	}
}

func (u *Releaser) Available(version string) (*Release, error) {
	currentVersion, err := semver.NewVersion(version)
	if err != nil {
		return nil, err
	}

	resp, err := u.Client.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", u.owner, u.repo))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.StatusCode)
		return nil, errors.New("Could not check for update")
	}

	releases := []Release{}

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, nil
	}

	name := releases[0].Name
	if strings.HasPrefix(name, "v") {
		name = name[1:]
	}

	if releaseVer, err := semver.NewVersion(name); err != nil {
		// could not parse version
		return nil, err
	} else if releaseVer.LessThan(*currentVersion) {
		return nil, nil
	} else if releaseVer.Equal(*currentVersion) {
		return nil, nil
	}

	return &releases[0], nil
}
