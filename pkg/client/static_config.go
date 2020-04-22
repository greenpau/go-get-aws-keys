package client

type StaticConfiguration struct {
	SamlResponseFile string `xml:"saml_response_file,attr" json:"saml_response_file" yaml:"saml_response_file"`
}
