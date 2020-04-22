package client

import (
	"encoding/base64"
)

type StateMachine struct {
	Metadata          SamlServiceMetadata
	AuthenticationURL string `xml:"auth_url,attr" json:"auth_url" yaml:"auth_url"`
	Saml              SamlStateMachine
}

type SamlServiceMetadata struct {
	Raw   []byte
	Plain string
	File  File
	URL   string
}

type SamlResponseAssertions struct {
	Raw   []byte
	Plain string
	File  File
}

type SamlStateMachine struct {
	Response   SamlResponse
	Attributes *SamlResponseData
	Assertions *SamlResponseAssertions
}

func (a *SamlResponseAssertions) GetPath() string {
	return a.File.Path
}

func (a *SamlResponseAssertions) GetEncoded() string {
	return base64.StdEncoding.EncodeToString(a.Raw)
}
