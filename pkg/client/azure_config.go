package client

type AzureConfiguration struct {
	TenantID      string `xml:"tenant_id,attr" json:"tenant_id" yaml:"tenant_id"`
	ApplicationID string `xml:"application_id,attr" json:"application_id" yaml:"application_id"`
}
