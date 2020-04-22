package client

type Configuration struct {
	Static   StaticConfiguration `xml:"static,attr" json:"static" yaml:"static"`
	Adfs     AdfsConfiguration   `xml:"adfs,attr" json:"adfs" yaml:"adfs"`
	Azure    AzureConfiguration  `xml:"azure,attr" json:"azure" yaml:"azure"`
	Aws      AwsConfiguration    `xml:"aws,attr" json:"aws" yaml:"aws"`
	Username string              `xml:"email,attr" json:"email" yaml:"email"`
	Password string              `xml:"password,attr" json:"password" yaml:"password"`
	Domain   string              `xml:"domain,attr" json:"domain" yaml:"domain"`
	File     File
}

type File struct {
	Dir  string
	Name string
	Path string
}
