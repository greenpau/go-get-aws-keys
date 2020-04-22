package client

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/url"
	"text/template"
	"time"
)

// AzureAuthnRequest is SAML AuthnRequest components.
type AzureAuthnRequest struct {
	URL           string
	ID            string
	TenantID      string
	ApplicationID string
	ConsumerURL   string
}

// GetAzureAuthnRequest returns Azure SAML Authen Request.
func (c *Client) GetAzureAuthnRequest() (*AzureAuthnRequest, error) {
	r := &AzureAuthnRequest{}
	tb := &bytes.Buffer{}
	// Create UUID for the request
	requestUUID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("Error generating UUID: %s", err)
	}
	// Create a text template for the request
	samlAuthRequestTemplate := `<samlp:AuthnRequest` +
		` xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"` +
		` xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"` +
		` ID="AWSSAML{{ .ID }}"` +
		` Version="2.0"` +
		` AssertionConsumerServiceURL="https://signin.aws.amazon.com/saml"` +
		` Destination="https://login.microsoftonline.com/` + c.Config.Azure.TenantID + `/saml2"` +
		` IssueInstant="{{ .Timestamp }}"` +
		` ProtocolBinding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST">` +
		`<saml:Issuer>{{ .Issuer }}</saml:Issuer>`

	if c.Config.Domain != "" {
		samlAuthRequestTemplate += `<samlp:Scoping>` +
			`<samlp:IDPList>` +
			`<samlp:IDPEntry ProviderID="https://` + c.Config.Domain + `" Name="` + c.Config.Domain + `" />` +
			`</samlp:IDPList>` +
			`</samlp:Scoping>`
	}

	/*
		samlAuthRequestTemplate += `<samlp:NameIDPolicy Format="urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress" />` +
		`<samlp:RequestedAuthnContext>` +
		`<saml:AuthnContextClassRef>urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport</saml:AuthnContextClassRef>` +
		`</samlp:RequestedAuthnContext>` +
	*/

	samlAuthRequestTemplate += `</samlp:AuthnRequest>`
	// Create a timestamp for the request
	e := time.Now().UTC()
	ns := e.Nanosecond()
	if ns != 0 {
		ns = int(ns / 100)
	}
	requestTimestamp := fmt.Sprintf(
		"%d-%02d-%02dT%02d:%02d:%02d.%07dZ",
		e.Year(), e.Month(), e.Day(), e.Hour(), e.Minute(), e.Second(), ns,
	)
	// Check whether Azure Application ID exists
	if c.Config.Azure.ApplicationID == "" {
		return nil, fmt.Errorf("application id is not set")
	}
	// Create parameters for the template
	p := SamlAuthRequestParams{
		ID:        requestUUID.String(),
		Timestamp: requestTimestamp,
		Issuer:    c.Config.Azure.ApplicationID,
	}
	// Apply the parameters to the template
	t, err := template.New("SamlAuthRequestContent").Parse(samlAuthRequestTemplate)
	if err != nil {
		return nil, err
	}
	if err = t.Execute(tb, p); err != nil {
		return nil, err
	}
	samlAuthRequest := tb.String()
	log.Debugf("SAML Authentication Request: %s", samlAuthRequest)
	// Create query parameters, SAML enflated base64-encoded string
	compressedSamlAuthRequest := &bytes.Buffer{}
	flater, err := flate.NewWriter(compressedSamlAuthRequest, flate.DefaultCompression)
	if err != nil {
		return nil, fmt.Errorf("Error compressing data: %s", err)
	}
	_, err = flater.Write([]byte(samlAuthRequest))
	if err != nil {
		return nil, fmt.Errorf("Error writing compressed data: %s", err)
	}
	flater.Close()
	log.Debugf("Compressed flated SAML AuthnRequest")
	encodedSamlAuthRequest := base64.StdEncoding.EncodeToString(compressedSamlAuthRequest.Bytes())
	log.Debugf("Base64-encoded, flated SAML AuthnRequest: %s", encodedSamlAuthRequest)

	//queryParameters := "?username=" + c.Config.Username + "&SAMLRequest=" + url.QueryEscape(encodedSamlAuthRequest)
	queryParameters := "?SAMLRequest=" + url.QueryEscape(encodedSamlAuthRequest)
	log.Debugf("URL: %s%s", c.Runtime.AuthenticationURL, queryParameters)
	if queryParameters == "?SAMLRequest=" {
		return nil, fmt.Errorf("empty query parameters")
	}

	/*
	   req, err := http.NewRequest("GET", c.Config.Adfs.AuthenticationURL+queryParameters, nil)
	   if err != nil {
	       return fmt.Errorf("Error creating http post: %s", err)
	   }
	   req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36")
	   resp, err := c.browser.Do(req)
	   if err != nil {
	       return fmt.Errorf("Error authenticating @ %s: %s", c.Config.Adfs.AuthenticationURL, err)
	   }
	   defer resp.Body.Close()
	   body, err := ioutil.ReadAll(resp.Body)
	   if err != nil {
	       return fmt.Errorf("Error reading response data from %s: %s", c.Config.Adfs.AuthenticationURL, err)
	   }
	   responseBody := string(body[:])

	   parsedResp, err := parseWebResponse(resp, responseBody)
	   if err != nil {
	       return fmt.Errorf("Error parsing response data from %s: %s", c.Config.Adfs.AuthenticationURL, err)
	   }
	   return fmt.Errorf("Azure AD authentication failed: %v", parsedResp)
	*/

	r.URL = c.Runtime.AuthenticationURL + queryParameters
	r.ID = requestUUID.String()
	r.TenantID = c.Config.Azure.TenantID
	r.ApplicationID = c.Config.Azure.ApplicationID
	r.ConsumerURL = "https://signin.aws.amazon.com/saml"
	return r, nil
}
