package pi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/hyperhq/client-go/tools/clientcmd"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
	"log"
)

var (
	Version = ""
	Commit  = ""
	Build   = ""
)

var upcktimePath = "cktime.json"

type updateTimeRecord struct {
	LastUpdate time.Time `json:"lastUpdate"`
}

type CheckUpdate struct {
	Config string
}

func NewCheckUpdate() *CheckUpdate {
	return &CheckUpdate{
		Config: fmt.Sprintf("%s/%s", clientcmd.RecommendedConfigDir, upcktimePath),
	}
}

func (c *CheckUpdate) ReadTime() time.Time {
	now := time.Now()
	p, err := os.Open(c.Config)
	if os.IsNotExist(err) {
		now = now.Add(-25 * time.Hour)
		c.WriteTime(now)
		return now
	}
	if err != nil {
		if os.Getenv("HYPER_DEBUG") == "true" {
			log.Fatalf("read lastUpdate time from %v error: %v", c.Config, err)
		}
		now = now.Add(-25 * time.Hour)
		return now
	}
	var update updateTimeRecord
	if err = json.NewDecoder(p).Decode(&update); err != nil {
		return now
	}
	return update.LastUpdate
}

func (c *CheckUpdate) WriteTime(t time.Time) bool {
	data, err := json.Marshal(updateTimeRecord{t})
	if err != nil {
		return false
	}
	err = ioutil.WriteFile(c.Config, data, 0600)
	if err != nil {
		if os.Getenv("HYPER_DEBUG") == "true" {
			log.Fatalf("write lastUpdate time to %v error: %v", c.Config, err)
		}
		return false
	}
	return true
}

func CheckRelease() {
	client := github.NewClient(nil)
	opt := &github.ListOptions{}
	var (
		releases []*github.RepositoryRelease
		err      error
		latest   string
	)
	if releases, _, err = client.Repositories.ListReleases(context.Background(), "hyperhq", "pi", opt); err != nil {
		glog.V(4).Info("failed to list repo from github")
	} else {
		for _, r := range releases {
			if *r.TagName == "latest" {
				latest = strings.TrimSpace(strings.Split(*r.Body, "\n")[0])
				if latest == Version {
					return
				} else {
					fmt.Printf("\nThere is a new version: %v\n", latest)
				}
				break
			}
		}
		for _, r := range releases {
			if latest == *r.TagName {
				preRelease := ""
				for _, a := range r.Assets {
					if *r.Prerelease {
						preRelease = "(Pre-release) "
					}
					fmt.Printf("- %v%v\n", preRelease, *a.BrowserDownloadURL)
				}
				return
			}
		}
	}
}
