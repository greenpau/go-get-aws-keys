package client

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"net"
	"net/url"
	"strings"
)

// AdfsAuthForm contains successful AWS STS service response.
type AdfsAuthForm struct {
	URL    string
	Host   string
	Port   string
	Fields map[string]string
}

// NewAdfsAuthFormFromString returns AdfsAuthForm instance from an input string.
func NewAdfsAuthFormFromString(s string) (*AdfsAuthForm, error) {
	return NewAdfsAuthFormFromBytes([]byte(s))
}

// NewAdfsAuthFormFromBytes returns AdfsAuthForm instance from an input byte array.
func NewAdfsAuthFormFromBytes(s []byte) (*AdfsAuthForm, error) {
	authForm := AdfsAuthForm{}
	r := bytes.NewReader(s)
	iterator := html.NewTokenizer(r)
	for {
		tt := iterator.Next()
		if tt == html.ErrorToken {
			break
		}
		if tt == html.StartTagToken {
			t := iterator.Token()
			isForm := t.Data == "form"
			if isForm {
				validAttrCount := 0
				for _, attr := range t.Attr {
					if attr.Key == "id" && attr.Val == "options" {
						validAttrCount++
					} else if attr.Key == "method" && attr.Val == "post" {
						validAttrCount++
					} else if attr.Key == "action" && strings.HasPrefix(attr.Val, "http") {
						validAttrCount++
						authForm.URL = attr.Val
					} else {
						continue
					}
				}
				if validAttrCount == 3 {
					u, err := url.Parse(authForm.URL)
					if err != nil {
						return nil, fmt.Errorf("Failed to parse URL: %s", authForm.URL)
					}
					host, port, _ := net.SplitHostPort(u.Host)
					authForm.Host = host
					authForm.Port = port
					return &authForm, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("Authentication form not found")
}
