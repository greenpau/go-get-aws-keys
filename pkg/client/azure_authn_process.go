package client

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// DoAzureAuthnRequestWithAdfs uses auto-accelleration feature to authenticate to IDP.
func (c *Client) DoAzureAuthnRequestWithAdfs(r *AzureAuthnRequest) error {
	if c.Config.Username == "" {
		return fmt.Errorf("No username found for authentication")
	}
	if c.Config.Password == "" {
		return fmt.Errorf("No password found for authentication")
	}
	// Step 1: Request goes to Azure and the expectation is that it
	// redirects the request to IdP login page.

	req, err := http.NewRequest("GET", r.URL, nil)
	if err != nil {
		return fmt.Errorf("Error creating http get request: %s", err)
	}
	resp, err := c.browser.Do(req)
	if err != nil {
		return fmt.Errorf("Error authenticating @ %s: %s", c.Runtime.AuthenticationURL, err)
	}
	defer resp.Body.Close()
	var responseBody string
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("Error reading response data from %s: %s", c.Runtime.AuthenticationURL, err)
	} else {
		responseBody = string(body[:])
	}
	authForm, err := NewAdfsAuthFormFromString(responseBody)
	if err != nil {
		return fmt.Errorf("Error parsing response: %s", err)
	}
	// Step 2: Once we get an authentication form, we post our credentials
	// to that form.
	log.Debugf("ADFS Authentication Form: %v", authForm)
	adfsFormEntries := url.Values{}
	adfsFormEntries.Add("UserName", c.Config.Username)
	adfsFormEntries.Add("Password", c.Config.Password)
	adfsFormEntries.Add("Kmsi", "true")
	adfsFormEntries.Add("AuthMethod", "FormsAuthentication")
	adfsFormData := strings.NewReader(adfsFormEntries.Encode())
	log.Debugf("ADFS Authentication URL: %s", authForm.URL)
	log.Debugf("ADFS form data: %v", adfsFormEntries)
	log.Debugf("ADFS form data (encoded): %v", adfsFormData)
	adfsReq, err := http.NewRequest("POST", authForm.URL, adfsFormData)
	if err != nil {
		return fmt.Errorf("Error creating http post request: %s", err)
	}
	adfsReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	adfsReq.Header.Add("Content-Length", strconv.Itoa(len(adfsFormEntries.Encode())))
	adfsResp, err := c.browser.Do(adfsReq)
	if err != nil {
		return fmt.Errorf("Error authenticating @ %s: %s", authForm.URL, err)
	}
	defer adfsResp.Body.Close()
	adfsRespBodyBytes, err := ioutil.ReadAll(adfsResp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response data from %s: %s", authForm.URL, err)
	}
	adfsRespBody := string(adfsRespBodyBytes[:])
	log.Debugf("ADFS responded with %s: %s", adfsResp.Status, adfsRespBody)
	if adfsResp.StatusCode != 200 {
		return fmt.Errorf("ADFS form-based authentication failed")
	}
	adfsAuthResponseForm, err := NewAdfsAuthResponseFormFromString(adfsRespBody)
	if err != nil {
		return fmt.Errorf("Error reading form data from %s: %s", authForm.URL, err)
	}
	log.Debugf("ADFS Authentication Response Form: %v", adfsAuthResponseForm)
	// Step 3: Upon successful authentication, ADFS outputs the form
	// redirecting us back to Azure. The form contains RequestSecurityTokenResponse
	// SAML request. We read the form data and resubmit it to Azure.
	azureTokenRequestFormEntries := url.Values{}
	for k, v := range adfsAuthResponseForm.Fields {
		log.Debugf("Azure POST field %s => %s", k, v)
		azureTokenRequestFormEntries.Set(k, v)
	}
	azureTokenRequestFormData := strings.NewReader(azureTokenRequestFormEntries.Encode())
	log.Debugf("Azure RequestSecurityToken URL:  %s", adfsAuthResponseForm.URL)
	log.Debugf("Azure RequestSecurityToken data: %v", azureTokenRequestFormEntries)
	log.Debugf("Azure RequestSecurityToken data(encoded): %v", azureTokenRequestFormData)
	azureTokenRequest, err := http.NewRequest("POST", adfsAuthResponseForm.URL, azureTokenRequestFormData)
	if err != nil {
		return fmt.Errorf("Error creating http post request for token request: %s", err)
	}
	azureTokenRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	azureTokenRequest.Header.Add("Content-Length", strconv.Itoa(len(azureTokenRequestFormEntries.Encode())))
	azureTokenRequestResponse, err := c.browser.Do(azureTokenRequest)
	if err != nil {
		return fmt.Errorf("Error submitting token request @ %s: %s", adfsAuthResponseForm.URL, err)
	}
	defer azureTokenRequestResponse.Body.Close()
	azureTokenRequestResponseBodyBytes, err := ioutil.ReadAll(azureTokenRequestResponse.Body)
	if err != nil {
		return fmt.Errorf("Error reading token request response data from %s: %s", adfsAuthResponseForm.URL, err)
	}
	azureTokenRequestResponseBody := string(azureTokenRequestResponseBodyBytes[:])
	log.Debugf("Azure responded with %s: %s", azureTokenRequestResponse.Status, azureTokenRequestResponseBody)
	if azureTokenRequestResponse.StatusCode != 200 {
		return fmt.Errorf("Azure authentication with ADFS form-based authentication response failed")
	}
	azureAuthResponseForm, err := NewAzureAuthResponseFormFromString(azureTokenRequestResponseBody)
	if err != nil {
		return fmt.Errorf("Error reading response form data from %s: %s", adfsAuthResponseForm.URL, err)
	}
	log.Debugf("Azure Authentication Response Form: %v", azureAuthResponseForm)
	// Step 4: Upon successful authentication with Azure using the ADFS response, we parse the form
	// and extract SAMLResponse value for the submission to AWS STS Endpoint.

	c.Runtime.Saml.Assertions = &SamlResponseAssertions{}

	c.Runtime.Saml.Assertions.Raw, err = base64.StdEncoding.DecodeString(azureAuthResponseForm.Fields["SAMLResponse"])
	if err != nil {
		return fmt.Errorf("Failed to decode SAMLResponse in Azure Authentication Response Form: %s", err)
	}
	c.Runtime.Saml.Assertions.Plain = string(c.Runtime.Saml.Assertions.Raw[:])

	if err := xml.Unmarshal(c.Runtime.Saml.Assertions.Raw, &c.Runtime.Saml.Response); err != nil {
		return fmt.Errorf("Failed to unmarshal SAMLResponse in Azure Authentication Response Form: %s", err)
	}
	if c.Runtime.Saml.Response.Assertion.AttributeStatement == nil {
		return fmt.Errorf("SAMLResponse in Azure Authentication Response Form does not contain attribute statements: %v", c.Runtime.Saml.Response.Assertion)
	}
	c.Runtime.Saml.Attributes, err = c.Runtime.Saml.Response.GetAttributes()
	if err != nil {
		return fmt.Errorf("Failed to get attributes from SAMLResponse in Azure Authentication Response Form: %s", err)
	}

	return nil
	//return fmt.Errorf("DoAzureAuthnRequestWithPost Pending Development")
}
