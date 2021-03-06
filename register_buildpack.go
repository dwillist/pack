package pack

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"github.com/buildpacks/pack/internal/buildpackage"
	"github.com/buildpacks/pack/internal/dist"
	"github.com/buildpacks/pack/internal/registry"
)

type RegisterBuildpackOptions struct {
	ImageName string
	Type      string
	URL       string
}

func (c *Client) RegisterBuildpack(ctx context.Context, opts RegisterBuildpackOptions) error {
	appImage, err := c.imageFetcher.Fetch(ctx, opts.ImageName, false, true)
	if err != nil {
		return err
	}

	var buildpackInfo dist.BuildpackInfo
	if _, err := dist.GetLabel(appImage, buildpackage.MetadataLabel, &buildpackInfo); err != nil {
		return err
	}

	namespace, name, err := parseID(buildpackInfo.ID)
	if err != nil {
		return err
	}

	id, err := appImage.Identifier()
	if err != nil {
		return err
	}

	buildpack := registry.Buildpack{
		Namespace: namespace,
		Name:      name,
		Version:   buildpackInfo.Version,
		Address:   id.String(),
		Yanked:    false,
	}

	issueURL, err := parseURL(opts.URL)
	if err != nil {
		return err
	}

	issue, err := registry.CreateGithubIssue(buildpack)
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Add("title", issue.Title)
	params.Add("body", issue.Body)
	issueURL.RawQuery = params.Encode()

	c.logger.Debugf("Open URL in browser: %s", issueURL)
	cmd, err := registry.CreateBrowserCmd(issueURL.String(), runtime.GOOS)
	if err != nil {
		return err
	}

	return cmd.Start()
}

func parseID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) < 2 {
		return "", "", errors.New("invalid id: does not contain a namespace")
	} else if len(parts) > 2 {
		return "", "", errors.New("invalid id: contains unexpected characters")
	}

	return parts[0], parts[1], nil
}

func parseURL(githubURL string) (*url.URL, error) {
	if githubURL == "" {
		return nil, errors.New("missing github URL")
	}
	return url.Parse(fmt.Sprintf("%s/issues/new", strings.TrimSuffix(githubURL, "/")))
}
