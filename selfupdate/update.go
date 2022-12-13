package selfupdate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/dgmorales/go-cli-selfupdate/logger"
	"github.com/google/go-github/v48/github"
	minioSelfUpdate "github.com/minio/selfupdate"
)

// NewClient creates a new GitHub client
func NewClient() *github.Client {
	return github.NewClient(nil)
}

// downloadAsset saves a GitHub Release Asset to a temporary file
func downloadAsset(i CliInfo, gh *github.Client) (filename string, err error) {
	var httpClient http.Client

	if i.assetName == "" || i.assetID == 0 || i.assetUrl == "" {
		return "", errors.New("in downloadAsset: github release asset information is unavailable")
	}

	filename = path.Join(logger.WorkDir, i.assetName)

	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("error creating file for release download: %w", err)
	}
	defer f.Close()

	data, redirectUrl, err := gh.Repositories.DownloadReleaseAsset(context.Background(),
		i.RepositoryOwner, i.RepositoryName, i.assetID, &httpClient)
	if redirectUrl != "" {
		return filename, fmt.Errorf("in downloadAsset: got unsupported asset redirect URL (%s)", redirectUrl)
	}
	if err != nil {
		return filename, fmt.Errorf("in downloadAsset: %w", err)
	}

	_, err = io.Copy(f, data)
	if err != nil {
		return filename, fmt.Errorf("in downloadAsset: %w", err)
	}

	return filename, nil
}

func DownloadAndApply(i CliInfo, gh *github.Client) error {
	filename, err := downloadAsset(i, gh)

	if err != nil {
		return err
	}

	log.Printf("Downloaded %s\n ...", filename)
	reader, closer, err := Uncompress(filename)
	if err != nil {
		if closer != nil {
			closer()
		}
		return err
	}

	err = minioSelfUpdate.Apply(reader, minioSelfUpdate.Options{})
	if err != nil {
		return err
	}

	return nil
}
