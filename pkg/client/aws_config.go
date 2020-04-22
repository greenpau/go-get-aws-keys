package client

type AwsConfigurationRole struct {
	AccountID     string `xml:"account_id,attr" json:"account_id" yaml:"account_id"`
	Name          string `xml:"role,attr" json:"role" yaml:"role"`
	ProfileName   string `xml:"profile_name,attr" json:"profile_name" yaml:"profile_name"`
	DefaultRegion string `xml:"region,attr" json:"region" yaml:"region"`
}

type AwsConfiguration struct {
	Roles             []*AwsConfigurationRole `xml:"roles,attr" json:"roles" yaml:"roles"`
	AuthenticationURL string                  `xml:"url,attr" json:"url" yaml:"url"`
}

type Aws struct {
	Credentials []*AwsCredentials
}
