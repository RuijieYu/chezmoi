package cmd

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/go-github/v42/github"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type gitHubData struct {
	sync.Mutex
	keysCache          map[string][]*github.Key
	latestReleaseCache map[string]map[string]*github.RepositoryRelease
}

func (c *Config) gitHubKeysTemplateFunc(user string) []*github.Key {
	c.gitHub.Lock()
	defer c.gitHub.Unlock()

	if keys, ok := c.gitHub.keysCache[user]; ok {
		return keys
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpClient, err := c.getHTTPClient()
	if err != nil {
		returnTemplateError(err)
		return nil
	}
	gitHubClient := newGitHubClient(ctx, httpClient)

	var allKeys []*github.Key
	opts := &github.ListOptions{
		PerPage: 100,
	}
	for {
		keys, resp, err := gitHubClient.Users.ListKeys(ctx, user, opts)
		if err != nil {
			returnTemplateError(err)
			return nil
		}
		allKeys = append(allKeys, keys...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if c.gitHub.keysCache == nil {
		c.gitHub.keysCache = make(map[string][]*github.Key)
	}
	c.gitHub.keysCache[user] = allKeys
	return allKeys
}

func (c *Config) gitHubLatestReleaseTemplateFunc(userRepo string) *github.RepositoryRelease {
	c.gitHub.Lock()
	defer c.gitHub.Unlock()

	user, repo := parseGitHubUserRepo(userRepo)

	if release := c.gitHub.latestReleaseCache[user][repo]; release != nil {
		return release
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpClient, err := c.getHTTPClient()
	if err != nil {
		returnTemplateError(err)
		return nil
	}
	gitHubClient := newGitHubClient(ctx, httpClient)

	release, _, err := gitHubClient.Repositories.GetLatestRelease(ctx, user, repo)
	if err != nil {
		returnTemplateError(err)
		return nil
	}

	if c.gitHub.latestReleaseCache == nil {
		c.gitHub.latestReleaseCache = make(map[string]map[string]*github.RepositoryRelease)
	}
	if c.gitHub.latestReleaseCache[user] == nil {
		c.gitHub.latestReleaseCache[user] = make(map[string]*github.RepositoryRelease)
	}
	c.gitHub.latestReleaseCache[user][repo] = release

	return release
}

func parseGitHubUserRepo(userRepo string) (string, string) {
	user, repo, ok := chezmoi.CutString(userRepo, "/")
	if !ok {
		returnTemplateError(fmt.Errorf("%s: not a user/repo", userRepo))
		return "", ""
	}
	return user, repo
}
