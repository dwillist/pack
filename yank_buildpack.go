package pack

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"github.com/buildpacks/pack/internal/style"

	"github.com/buildpacks/pack/internal/registry"
)

type YankBuildpackOptions struct {
	ID      string
	Version string
	Type    string
	URL     string
	Yanked  bool
}

func (c *Client) YankBuildpack(opts YankBuildpackOptions) error {
	namespace, name, err := parseNamespaceName(opts.ID)
	if err != nil {
		return err
	}
	issueURL, err := registry.GetIssueURL(opts.URL)
	if err != nil {
		return err
	}

	buildpack := registry.Buildpack{
		Namespace: namespace,
		Name:      name,
		Version:   opts.Version,
		Yanked:    opts.Yanked,
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

func parseNamespaceName(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid id %s does not contain a namespace", style.Symbol(id))
	} else if len(parts) > 2 {
		return "", "", fmt.Errorf("invalid id %s contains unexpected characters", style.Symbol(id))
	}

	return parts[0], parts[1], nil
}
