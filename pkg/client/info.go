package client

import (
	"fmt"
	"strings"
)

// Info holds information about the package
type Info struct {
	Name          string
	Version       string
	Description   string
	Documentation string
	Git           GitInfo
	Build         BuildInfo
}

// GitInfo holds information Git-related information.
type GitInfo struct {
	Branch string
	Commit string
}

// BuildInfo  holds information build-related information.
type BuildInfo struct {
	OperatingSystem string
	Architecture    string
	User            string
	Date            string
}

// GetVersionInfo returns version information
func (c *Client) GetVersionInfo() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s %s", c.Info.Name, c.Info.Version))
	if c.Info.Git.Branch != "" {
		sb.WriteString(fmt.Sprintf(", branch: %s", c.Info.Git.Branch))
	}
	if c.Info.Git.Commit != "" {
		sb.WriteString(fmt.Sprintf(", commit: %s", c.Info.Git.Commit))
	}
	if c.Info.Build.User != "" && c.Info.Build.Date != "" {
		sb.WriteString(fmt.Sprintf(", build on %s by %s",
			c.Info.Build.Date, c.Info.Build.User,
		))
		if c.Info.Build.OperatingSystem != "" && c.Info.Build.Architecture != "" {
			sb.WriteString(
				fmt.Sprintf(" for %s/%s",
					c.Info.Build.OperatingSystem, c.Info.Build.Architecture,
				))
		}
	}
	return sb.String()
}
