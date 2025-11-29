package github

import (
	"net/http"

	"github.com/google/go-github/v77/github"
)

type Client struct {
	Owner  string
	Name   string
	github *github.Client
}

func NewClient(token string, repo string, httpClient *http.Client) (*Client, error) {
	owner, name, err := ParseRepository(repo)
	if err != nil {
		return nil, err
	}

	var ghClient *github.Client
	if httpClient != nil {
		ghClient = github.NewClient(httpClient)
	} else {
		ghClient = github.NewClient(nil)
	}

	if token != "" {
		ghClient = ghClient.WithAuthToken(token)
	}

	return &Client{Owner: owner, Name: name, github: ghClient}, nil
}
