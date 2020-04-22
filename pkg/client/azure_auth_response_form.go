package client

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"net"
	"net/url"
	"strings"
)

// AzureAuthResponseForm contains successful AWS STS service response.
type AzureAuthResponseForm struct {
	URL    string
	Host   string
	Port   string
	Fields map[string]string
}

// NewAzureAuthResponseFormFromString returns AzureAuthResponseForm instance from an input string.
func NewAzureAuthResponseFormFromString(s string) (*AzureAuthResponseForm, error) {
	return NewAzureAuthResponseFormFromBytes([]byte(s))
}

// NewAzureAuthResponseFormFromBytes returns AzureAuthResponseForm instance from an input byte array.
func NewAzureAuthResponseFormFromBytes(s []byte) (*AzureAuthResponseForm, error) {
	authResponseForm := AzureAuthResponseForm{}
	authResponseFormFields := map[string]string{}
	r := bytes.NewReader(s)
	iterator := html.NewTokenizer(r)
	isProcessingForm := false
	for {
		tt := iterator.Next()
		if tt == html.ErrorToken {
			break
		}
		if tt == html.StartTagToken {
			t := iterator.Token()
			isForm := t.Data == "form"
			if isForm {
				isProcessingForm = true
				for _, attr := range t.Attr {
					if attr.Key == "action" && strings.HasPrefix(attr.Val, "http") {
						authResponseForm.URL = attr.Val
						break
					}
				}
				if authResponseForm.URL != "" {
					u, err := url.Parse(authResponseForm.URL)
					if err != nil {
						return nil, fmt.Errorf("Failed to parse URL: %s", authResponseForm.URL)
					}
					if strings.Contains(u.Host, ":") {
						if host, port, err := net.SplitHostPort(u.Host); err != nil {
							return nil, fmt.Errorf("Failed to parse host and port: %s", authResponseForm.URL)
						} else {
							authResponseForm.Host = host
							authResponseForm.Port = port
						}
					} else {
						authResponseForm.Host = u.Host
					}
				}
			}
			continue
		}

		if isProcessingForm && tt == html.EndTagToken {
			// Reached the end of the form
			break
		}

		if tt == html.SelfClosingTagToken {
			t := iterator.Token()
			var k, v string
			//var err error
			for _, attr := range t.Attr {
				if attr.Key == "name" {
					k = attr.Val
				}
				if attr.Key == "value" {
					v = attr.Val
				}
			}
			if k != "" && v != "" {
				authResponseFormFields[k] = v
			}
		}
	}
	if authResponseForm.URL == "" {
		return nil, fmt.Errorf("The Azure authentication response does not contain a form with SAMLResponse input field")
	}

	for _, k := range []string{"SAMLResponse"} {
		if _, exists := authResponseFormFields[k]; !exists {
			return nil, fmt.Errorf("The Azure authentication response form does not contain '%s' field", k)
		}
	}

	authResponseForm.Fields = authResponseFormFields
	return &authResponseForm, nil
}
