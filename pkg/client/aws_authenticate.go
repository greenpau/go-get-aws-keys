package client

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// AssumeRoleWithSaml makes AWS API call to STS service and asks for
//temporary credentials.
func (c *Client) AssumeRoleWithSaml() error {
	roles := c.GetRequestedAwsRoles()
	if len(roles) == 0 {
		return fmt.Errorf("The available AWS roles do no match any of the requested AWS roles")
	}
	for _, role := range roles {
		assertions := c.Runtime.Saml.Assertions
		log.Debugf("ADFS-provided SAML Assertions: %s\n\n", assertions.Plain)
		encodedAssertions := assertions.GetEncoded()
		keyValuePairs := url.Values{}
		keyValuePairs.Add("Version", "2011-06-15")
		keyValuePairs.Add("Action", "AssumeRoleWithSAML")
		keyValuePairs.Add("RoleArn", role.RoleARN)
		keyValuePairs.Add("PrincipalArn", role.IdentityProviderARN)
		keyValuePairs.Add("SAMLAssertion", encodedAssertions)
		postData := strings.NewReader(keyValuePairs.Encode())
		log.Debugf("AWS STS Authentication URL: %s", c.Config.Aws.AuthenticationURL)
		req, err := http.NewRequest("POST", c.Config.Aws.AuthenticationURL, postData)
		if err != nil {
			log.Errorf("Error creating http post when assuming AWS role: %s", err)
			continue
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(keyValuePairs.Encode())))
		req.Header.Add("Accept", "application/json")
		resp, err := c.browser.Do(req)
		if err != nil {
			log.Errorf("Error authenticating @ %s when assuming AWS role: %s", c.Config.Aws.AuthenticationURL, err)
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("Error reading response data from %s when assuming AWS role: %s", c.Config.Aws.AuthenticationURL, err)
			continue
		}
		awsStsResponse, err := NewAwsStsResponseFromBytes(body)
		if err != nil {
			log.Errorf("Error decoding STS response when assuming AWS role: %s", err)
			continue
		}
		log.Debugf("Received AWS STS Response: %v", awsStsResponse)
		awsCredentials, err := NewAwsCredentialsFromStsResponse(awsStsResponse)
		if err != nil {
			log.Errorf("Error creating AWS credentials when assuming AWS role: %s", err)
			continue
		}
		log.Debugf("The AWS Credentials are: %v", awsCredentials)
		awsCredentials.ProfileName = role.ProfileName
		awsCredentials.DefaultRegion = role.DefaultRegion
		c.Aws.Credentials = append(c.Aws.Credentials, awsCredentials)
	}
	return nil
}
