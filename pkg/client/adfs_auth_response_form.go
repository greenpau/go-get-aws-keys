package client

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"net"
	"net/url"
	"sort"
	"strings"
)

// AdfsAuthResponseForm contains successful AWS STS service response.
type AdfsAuthResponseForm struct {
	URL    string
	Host   string
	Port   string
	Fields map[string]string
}

// NewAdfsAuthResponseFormFromString returns AdfsAuthResponseForm instance from an input string.
func NewAdfsAuthResponseFormFromString(s string) (*AdfsAuthResponseForm, error) {
	return NewAdfsAuthResponseFormFromBytes([]byte(s))
}

// NewAdfsAuthResponseFormFromBytes returns AdfsAuthResponseForm instance from an input byte array.
func NewAdfsAuthResponseFormFromBytes(s []byte) (*AdfsAuthResponseForm, error) {
	authResponseForm := AdfsAuthResponseForm{}
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
					host, port, _ := net.SplitHostPort(u.Host)
					authResponseForm.Host = host
					authResponseForm.Port = port
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
					//v, err = url.QueryUnescape(attr.Val)
					//if err != nil {
					//	return nil, fmt.Errorf("Failed to unescape form input value: %s, %s", err, attr.Val)
					//}
				}
			}

			if k != "" && v != "" {
				authResponseFormFields[k] = v
			}
		}
	}
	if authResponseForm.URL == "" {
		return nil, fmt.Errorf("The ADFS authentication response does not contain a form with RequestSecurityTokenResponse")
	}

	for _, k := range []string{"wa", "wctx", "wresult"} {
		if _, exists := authResponseFormFields[k]; !exists {
			return nil, fmt.Errorf("The ADFS authentication response form does not contain '%s' field", k)
		}
	}

	// Remove spaces from X509Certificate and ds:SignatureValue
	keyRunes := []rune(authResponseFormFields["wresult"])
	emptyOffsets := []int{}
	for _, keyName := range []string{"X509Certificate", "ds:SignatureValue", "ds:DigestValue"} {
		keyBeginOffset := strings.Index(authResponseFormFields["wresult"], keyName)
		if keyBeginOffset < 0 {
			continue
		}
		keyEndOffset := strings.Index(authResponseFormFields["wresult"], "/"+keyName)
		if keyEndOffset < 0 {
			continue
		}
		keyValue := string(keyRunes[keyBeginOffset:keyEndOffset])
		keyValueRunes := []rune(keyValue)
		for i, r := range keyValueRunes {
			if r == ' ' {
				emptyOffsets = append(emptyOffsets, keyBeginOffset+i)
			}
		}
	}
	if len(emptyOffsets) > 0 {
		sort.Sort(sort.Reverse(sort.IntSlice(emptyOffsets)))
		for _, offset := range emptyOffsets {
			keyRunes = append(keyRunes[:offset], keyRunes[offset+1:]...)
		}
		authResponseFormFields["wresult"] = string(keyRunes)
	}
	authResponseFormFields["LoginOptions"] = "1"
	authResponseForm.Fields = authResponseFormFields
	return &authResponseForm, nil
}
