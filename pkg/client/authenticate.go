package client

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type SamlAuthRequestParams struct {
	ID        string
	Issuer    string
	Timestamp string
}

// GetSamlAssertions requests SAML assertions either from
// ADFS instance, Azure AD, or local file.
func (c *Client) GetSamlAssertions() error {
	if err := c.GetAuthenticationURL(); err != nil {
		return err
	}

	if c.Config.Azure.TenantID != "" {
		if err := c.AuthenticateWithAzure(); err != nil {
			return err
		}
	} else if c.Config.Adfs.Hostname != "" {
		if err := c.AuthenticateWithAdfs(); err != nil {
			return err
		}
	}
	if err := c.OutputCurrentState(); err != nil {
		return err
	}
	if err := c.IsSamlAssertionValid(); err != nil {
		return err
	}
	if err := c.IsAwsRoleAvailable(); err == nil {
		return err
	}
	return nil
}

func (c *Client) IsSamlAssertionValid() error {
	// TODO: check the fields
	// resp.Aws.AuthenticateByTimestamp = r.Assertion.Subject.Confirmation.Data.NotOnOrAfter
	// resp.Aws.SessionStartTimestamp = r.Assertion.Conditions.NotBefore
	// if onOrAfter {
	//    return fmt.Errorf("SAML Assertion is not valid due to NotOnOrAfter for the assertion being met: %s",  c.Runtime.Saml.Attributes.Aws.Conditions.NotOnOrAfter);
	// }
	// if notBefore {
	//    return fmt.Errorf("SAML Assertion is invalid due to NotBefore for the session being met: %s",  c.Runtime.Saml.Attributes.Aws.Conditions.NotBefore)
	// }
	return nil
}

func (c *Client) GetRequestedAwsRoles() []*AwsRole {
	roles := []*AwsRole{}
	for _, role := range c.Runtime.Saml.Attributes.Aws.Roles {
		for _, configRole := range c.Config.Aws.Roles {
			if configRole.AccountID != role.AccountID {
				continue
			}
			if configRole.Name != role.Name {
				continue
			}
			role.ProfileName = configRole.ProfileName
			role.DefaultRegion = configRole.DefaultRegion
			roles = append(roles, role)
			break
		}
	}
	return roles
}

func (c *Client) IsAwsRoleAvailable() error {
	for _, configRole := range c.Config.Aws.Roles {
		isRoleFound := false
		for _, runtimeRole := range c.Runtime.Saml.Attributes.Aws.Roles {
			if configRole.AccountID != runtimeRole.AccountID {
				continue
			}
			if configRole.Name != runtimeRole.Name {
				continue
			}
			log.Debugf("The requested IAM Role %s on account ID %s was issued by ADFS", runtimeRole.Name, runtimeRole.AccountID)
			isRoleFound = true
			break
		}
		if !isRoleFound {
			return fmt.Errorf("The requested IAM Role %s on account ID %s was NOT issued by ADFS", configRole.Name, configRole.AccountID)

		}
	}
	return nil
}

func (c *Client) OutputCurrentState() error {
	if len(c.Runtime.Saml.Attributes.Aws.Roles) < 1 {
		return fmt.Errorf("SAML Assertions about AWS Roles not found")
	}
	if c.Runtime.Saml.Attributes.Aws.SessionName == "" {
		return fmt.Errorf("SAML Assertions about AWS Role Session Name not found")
	}
	log.Debugf("ADFS authorized AWS Session Name: %s", c.Runtime.Saml.Attributes.Aws.SessionName)
	log.Debug("ADFS authorized AWS IAM Roles")
	for _, role := range c.Runtime.Saml.Attributes.Aws.Roles {
		log.Debugf("  - %s on account ID %s", role.Name, role.AccountID)
	}
	if c.Runtime.Saml.Attributes.Aws.SessionDuration > 0 {
		log.Debugf("ADFS authorized AWS session duration: %d", c.Runtime.Saml.Attributes.Aws.SessionDuration)
	}
	if !c.Runtime.Saml.Attributes.Aws.AuthenticateByTimestamp.IsZero() {
		log.Debugf("ADFS expects authentication by: %s", c.Runtime.Saml.Attributes.Aws.AuthenticateByTimestamp)
	}
	log.Debugf("ADFS limits the session between %s and %s", c.Runtime.Saml.Attributes.Aws.SessionStartTimestamp, c.Runtime.Saml.Attributes.Aws.SessionEndTimestamp)
	if len(c.Runtime.Saml.Attributes.Claims) > 0 {
		log.Debugf("ADFS responded with the following SAML Claims:")
		for _, claim := range c.Runtime.Saml.Attributes.Claims {
			log.Debugf("  - %s => %s", claim.Type, claim.Value)
		}
	}
	if c.Runtime.Saml.Attributes.Issuer != "" {
		log.Debugf("SAML Issuer: %s", c.Runtime.Saml.Attributes.Issuer)
	}
	log.Debugf("ADFS Success: %t", c.Runtime.Saml.Attributes.Success)
	return nil
}

// AuthenticateWithAdfs authenticates to ADFS and receives SAML assertions back.
func (c *Client) AuthenticateWithAdfs() error {
	formData, err := c.GetAdfsAuthenticationRequestBody()
	if err != nil {
		return err
	}
	resp, err := http.PostForm(c.Runtime.AuthenticationURL, formData)
	if err != nil {
		return fmt.Errorf("Error authenticating @ %s: %s", c.Runtime.AuthenticationURL, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response data from %s: %s", c.Runtime.AuthenticationURL, err)
	}
	responseBody := string(body[:])
	log.Debugf("ADFS responded with: %s", responseBody)
	return nil
}

// AuthenticateWithAzure authenticates to Azure AD and receives SAML assertions back.
func (c *Client) AuthenticateWithAzure() error {
	r, err := c.GetAzureAuthnRequest()
	if err != nil {
		return err
	}
	err = c.DoAzureAuthnRequestWithAdfs(r)
	if err != nil {
		return err
	}
	return nil
	//return fmt.Errorf("Pending Implementation")
}
