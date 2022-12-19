package gh

import (
	"context"
	"os"

	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

func NewClient() (*github.Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc), nil
}
