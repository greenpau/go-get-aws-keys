package client

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
)

// AwsCredentials holds raw AWS STS response.
type AwsCredentials struct {
	Raw             *AwsStsResponse
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
	ProfileName     string
	DefaultRegion   string
}

// IsValid check whether the credentials contain mandatory keys.
func (c *AwsCredentials) IsValid() error {
	if c.AccessKeyId == "" {
		return fmt.Errorf("AccessKeyId is empty")
	}
	if c.SecretAccessKey == "" {
		return fmt.Errorf("SecretAccessKey is empty")
	}
	if c.SessionToken == "" {
		return fmt.Errorf("SessionToken is empty")
	}
	return nil
}

// NewAwsCredentialsFromStsResponse return AwsCredentials from AwsStsResponse.
func NewAwsCredentialsFromStsResponse(resp *AwsStsResponse) (*AwsCredentials, error) {
	a := &AwsCredentials{
		Raw:             resp,
		AccessKeyId:     resp.Credentials.AccessKeyId,
		SecretAccessKey: resp.Credentials.SecretAccessKey,
		SessionToken:    resp.Credentials.SessionToken,
	}
	if err := a.IsValid(); err != nil {
		return nil, err
	}
	return a, nil
}

// WriteCredentialsFile writes the credentials to a file i.e. `.aws/credentials`.
// The function takes in a file path and a profile name. It creates a profile
// definition with the supplied namd and adds `aws_access_key_id`,
// `aws_secret_access_key`, and `aws_session_token` to the profile.
// If the profile exists, it overwrites.
func (c *AwsCredentials) WriteCredentialsFile(fp string) error {
	if err := c.IsValid(); err != nil {
		return err
	}
	fp = ExpandFilePath(fp)

	ff := os.O_APPEND | os.O_WRONLY
	isFileExists := true
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		isFileExists = false
		ff = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	}

	if isFileExists {
		var esb strings.Builder
		isWriter := true
		fh, err := os.Open(fp)
		if err != nil {
			return fmt.Errorf("Erred opening existing file %s: %s", fp, err)
		}
		defer fh.Close()
		scanner := bufio.NewScanner(fh)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "["+c.ProfileName+"]") {
				// Found the reference to a section with the target
				// profile name
				isWriter = false
			} else if strings.HasPrefix(scanner.Text(), "[") {
				// Found the reference to a section not containing the
				// target profile name
				isWriter = true
			}
			if isWriter {
				esb.WriteString(scanner.Text())
				esb.WriteRune('\n')
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("Erred reading existing file %s: %s", fp, err)
		}
		if err := ioutil.WriteFile(fp, []byte(strings.TrimLeft(esb.String(), "\n")), 0600); err != nil {
			return err
		}
	}

	var sb strings.Builder
	if isFileExists {
		sb.WriteRune('\n')
	}
	sb.WriteString(fmt.Sprintf("[%s]\n", c.ProfileName))
	sb.WriteString(fmt.Sprintf("# Assumed Role ID: %s\n", c.Raw.AssumedRoleUser.AssumedRoleId))
	sb.WriteString(fmt.Sprintf("# Assumed Role ARN: %s\n", c.Raw.AssumedRoleUser.Arn))
	sb.WriteString(fmt.Sprintf("region=%s\n", c.DefaultRegion))
	sb.WriteString(fmt.Sprintf("aws_access_key_id=%s\n", c.AccessKeyId))
	sb.WriteString(fmt.Sprintf("aws_secret_access_key=%s\n", c.SecretAccessKey))
	sb.WriteString(fmt.Sprintf("aws_session_token=%s\n", c.SessionToken))
	fh, err := os.OpenFile(fp, ff, 0600)
	if err != nil {
		return fmt.Errorf("Erred opening %s (existing: %t) for writing %s profile: %s", fp, isFileExists, c.ProfileName, err)
	}
	defer fh.Close()
	_, err = fh.WriteString(sb.String())
	if err != nil {
		return fmt.Errorf("Erred writing %s profile to %s (existing: %t): %s", c.ProfileName, fp, isFileExists, err)
	}
	log.Infof("Added %s aws credentials profile to %s", c.ProfileName, fp)
	return nil
}

// WriteEnvVarsFile writes an environment variables file which
// exports `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and
// `AWS_SESSION_TOKEN` environment variables.
func (c *AwsCredentials) WriteEnvVarsFile(fp string) error {
	if err := c.IsValid(); err != nil {
		return err
	}
	var sb strings.Builder
	exportWord := "export"
	sep := "="
	commentWord := "#"
	if runtime.GOOS == "windows" {
		exportWord = "SETX"
		sep = " "
		commentWord = "REM"
	}
	sb.WriteString(fmt.Sprintf("%s Assumed Role ID: %s\n", commentWord, c.Raw.AssumedRoleUser.AssumedRoleId))
	sb.WriteString(fmt.Sprintf("%s Assumed Role ARN: %s\n", commentWord, c.Raw.AssumedRoleUser.Arn))
	sb.WriteString(fmt.Sprintf("%s AWS_DEFAULT_REGION%s%s\n", exportWord, sep, c.DefaultRegion))
	sb.WriteString(fmt.Sprintf("%s AWS_ACCESS_KEY_ID%s%s\n", exportWord, sep, c.AccessKeyId))
	sb.WriteString(fmt.Sprintf("%s AWS_SECRET_ACCESS_KEY%s%s\n", exportWord, sep, c.SecretAccessKey))
	sb.WriteString(fmt.Sprintf("%s AWS_SESSION_TOKEN%s%s\n", exportWord, sep, c.SessionToken))
	if err := ioutil.WriteFile(fp, []byte(sb.String()), 0600); err != nil {
		return fmt.Errorf("Erred writing environment variables to %s: %s", fp, err)
	}
	log.Debugf("Wrote environment variables to %s", fp)
	return nil
}
