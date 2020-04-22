// Package client implements obtaining AWS STS Tokens by authenticating
// to ADFS (e.g. Azure AD) and passing the received SAML Claims to AWS.
package client

import (
	"bufio"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	appName              = "go-get-aws-keys"
	appVersion           = "[untracked]"
	appDescription       = "Obtain AWS STS Tokens by authenticating to Azure AD or ADFS and passing SAML Claims to AWS."
	appDocumentation     = "https://github.com/greenpau/go-get-aws-keys/"
	gitBranch            string
	gitCommit            string
	buildOperatingSystem string
	buildArchitecture    string
	buildUser            string
	buildDate            string
)

// Client is an instance of the compliance auditing utility for AWS.
type Client struct {
	sync.Mutex
	browser *http.Client
	Name    string
	Config  Configuration
	Runtime StateMachine
	Info    Info
	Aws     Aws
}

func (c *Client) init() {
	c.Info = Info{}
	if c.browser == nil {
		cj, err := cookiejar.New(nil)
		if err != nil {
			log.Errorf("Failed to create a cookie jar: %s", err)
			return
		}
		var tr = &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		}
		c.browser = &http.Client{
			Jar:       cj,
			Timeout:   time.Second * 10,
			Transport: tr,
		}
	}
	c.Info.Name = appName
	c.Info.Version = appVersion
	c.Info.Description = appDescription
	c.Info.Documentation = appDocumentation
	c.Info.Git.Branch = gitBranch
	c.Info.Git.Commit = gitCommit
	c.Info.Build.User = buildUser
	c.Info.Build.Date = buildDate
	c.Info.Build.OperatingSystem = buildOperatingSystem
	c.Info.Build.Architecture = buildArchitecture
	return
}

// New returns an instance of Client.
func New() *Client {
	c := &Client{
		Name: "go-get-aws-keys",
	}
	c.init()
	return c
}

// RequestAwsRole sets the desired IAM role name on AWS account to assume,
// together with a default region and profile name in AWS credentials file.
func (c *Client) RequestAwsRole(reqRole map[string]string) error {
	for _, k := range []string{"account_id", "name"} {
		if _, exists := reqRole[k]; !exists {
			return fmt.Errorf("The requested AWS role does not contain '%s' field", k)
		}
	}

	log.Debugf("Requested AWS Role %s on Account ID %s", reqRole["name"], reqRole["account_id"])

	role := &AwsConfigurationRole{
		AccountID: reqRole["account_id"],
		Name:      reqRole["name"],
	}

	if _, exists := reqRole["profile_name"]; exists {
		role.ProfileName = reqRole["profile_name"]
	}

	if _, exists := reqRole["region"]; exists {
		role.DefaultRegion = reqRole["region"]
	}

	c.Config.Aws.Roles = append(c.Config.Aws.Roles, role)
	return c.UpdateAwsRoles()
}

// UpdateAwsRoles iterates over the existing roles and throws an error
// when the role map is non-compliant.
func (c *Client) UpdateAwsRoles() error {
	if c.Config.Aws.Roles == nil {
		return nil
	}
	for _, role := range c.Config.Aws.Roles {
		if role.AccountID == "" {
			return fmt.Errorf("The requested AWS role does not contain 'account_id' field")
		}
		if role.Name == "" {
			return fmt.Errorf("The requested AWS role does not contain 'role' field")
		}
		if role.ProfileName == "" {
			role.ProfileName = "ggk-" + role.AccountID + "-" + role.Name
		}
		if role.DefaultRegion == "" {
			role.DefaultRegion = "us-east-1"
		}
	}
	c.Config.Aws.AuthenticationURL = "https://sts.amazonaws.com/"
	return nil
}

// SetAzureTenantID sets the tenant ID for Azure ADFS integration.
func (c *Client) SetAzureTenantID(s string) error {
	if s == "" {
		if c.Config.Azure.TenantID == "" {
			return fmt.Errorf("empty Azure Tenant ID")
		}
	} else {
		c.Config.Azure.TenantID = s
	}
	log.Debugf("Azure Tenant ID: %s", c.Config.Azure.TenantID)
	return nil
}

// SetAdfsHostname sets the hostname for enterprise ADFS instance.
func (c *Client) SetAdfsHostname(s string) error {
	if s == "" {
		if c.Config.Adfs.Hostname == "" {
			return fmt.Errorf("empty hostname for enterprise ADFS instance")
		}
	} else {
		c.Config.Adfs.Hostname = s
	}
	log.Debugf("Enterprise ADFS instance hostname: %s", c.Config.Adfs.Hostname)
	return nil
}

// SetAzureApplicationID sets the AWS Application ID for Azure ADFS integration.
func (c *Client) SetAzureApplicationID(s string) error {
	if s == "" {
		if c.Config.Azure.ApplicationID == "" {
			return fmt.Errorf("empty Azure AWS Application ID")
		}
	} else {
		c.Config.Azure.ApplicationID = s
	}
	log.Debugf("Azure Application ID for AWS: %s", c.Config.Azure.ApplicationID)
	return nil
}

// SetUsername sets username for ADFS requests.
func (c *Client) SetUsername(s string) error {
	if s == "" {
		if c.Config.Username == "" {
			return fmt.Errorf("empty username")
		}
	} else {
		c.Config.Username = s
	}
	log.Debugf("Username: %s", c.Config.Username)
	if c.Config.Azure.TenantID != "" {
		components := strings.Split(c.Config.Username, "@")
		if len(components) != 2 {
			return fmt.Errorf("username must be in email format")
		}
		c.Config.Domain = components[1]
		log.Debugf("Domain: %s", c.Config.Domain)
	}
	return nil
}

// SetPassword sets password for ADFS requests.
func (c *Client) SetPassword(s string) error {
	if s == "" {
		if c.Config.Password == "" {
			return fmt.Errorf("empty ADFS password")
		}
	} else {
		c.Config.Password = s
	}
	return nil
}

// SetStaticSamlResponseFile sets the path to the file with ADFS SAML Response.
func (c *Client) SetStaticSamlResponseFile(s string) error {
	if c.Config.Static.SamlResponseFile != "" {
		return nil
	}
	if s == "" {
		return fmt.Errorf("Empty ADFS token file")
	}
	if strings.HasPrefix(s, "~/") {
		usr, err := user.Current()
		if err != nil {
			return err
		}
		s = strings.Replace(s, "~", usr.HomeDir, 1)
	}
	assertions := &SamlResponseAssertions{}
	assertions.File.Dir = filepath.Dir(s)
	assertions.File.Name = filepath.Base(s)
	assertions.File.Path = path.Join(assertions.File.Dir, assertions.File.Name)
	c.Config.Static.SamlResponseFile = assertions.File.Path
	c.Runtime.Saml.Assertions = assertions

	if err := c.ReadStaticSamlResponseFile(); err != nil {
		return fmt.Errorf("Error while reading ADFS token file: %s", err)
	}
	log.Debugf("ADFS Token file (with SAML Response): %s", c.Config.Static.SamlResponseFile)
	return nil
}

// SetConfigFile sets the name and directory of the configuration file.
func (c *Client) SetConfigFile(s string) error {
	if s == "" {
		return nil
	}
	c.Config.File.Dir = filepath.Dir(s)
	c.Config.File.Name = filepath.Base(s)
	c.Config.File.Path = path.Join(c.Config.File.Dir, c.Config.File.Name)
	if c.Config.File.Dir == "" {
		return nil
	}
	c.Runtime.Metadata.File.Dir = c.Config.File.Dir
	if c.Config.Azure.TenantID != "" {
		c.Runtime.Metadata.File.Name = "azure." +
			c.Config.Azure.TenantID + "." +
			c.Config.Azure.ApplicationID + ".metadata.xml"
	}
	if c.Config.Adfs.Hostname != "" {
		c.Runtime.Metadata.File.Name = "adfs.enterprise." + c.Config.Adfs.Hostname + ".metadata.xml"

	}
	if c.Config.Azure.TenantID != "" || c.Config.Adfs.Hostname != "" {
		c.Runtime.Metadata.File.Path = path.Join(c.Runtime.Metadata.File.Dir, c.Runtime.Metadata.File.Name)
	}
	return nil
}

// IsMetadataNeeded returns true when metadata is not necessary, e.g.
// when SAML Response is available.
func (c *Client) IsMetadataNeeded() bool {
	if c.Config.Static.SamlResponseFile != "" {
		return false
	}
	return true
}

// IsMetadataExists checks whether metadata file exists
func (c *Client) IsMetadataExists() bool {
	if c.Config.Azure.TenantID != "" {
		c.Runtime.Metadata.URL = "https://login.microsoftonline.com/" +
			c.Config.Azure.TenantID +
			"/FederationMetadata/2007-06/FederationMetadata.xml?appid=" +
			c.Config.Azure.ApplicationID
	} else {
		c.Runtime.Metadata.URL = "https://" + c.Config.Adfs.Hostname + "/FederationMetadata/2007-06/FederationMetadata.xml"
	}
	if c.Runtime.Metadata.File.Path == "" {
		return false
	}
	if _, err := os.Stat(c.Runtime.Metadata.File.Path); os.IsNotExist(err) {
		return false
	}
	return true
}

// ReadMetadataFromFile reads ADFS metadata from a file.
func (c *Client) ReadMetadataFromFile() error {
	fh, err := os.Open(c.Runtime.Metadata.File.Path)
	if err != nil {
		return err
	}
	defer fh.Close()
	c.Runtime.Metadata.Raw, err = ioutil.ReadAll(fh)
	if err != nil {
		return err
	}
	c.Runtime.Metadata.Plain = string(c.Runtime.Metadata.Raw[:])
	return nil
}

// WriteMetadataFile writes a metadata file to the directory of the
// configuration file
func (c *Client) WriteMetadataToFile() error {
	if c.IsMetadataExists() {
		return nil
	}
	fh, err := os.Create(c.Runtime.Metadata.File.Path)
	if err != nil {
		return err
	}
	defer fh.Close()
	_, err = io.WriteString(fh, c.Runtime.Metadata.Plain)
	if err != nil {
		return err
	}
	return fh.Sync()
	return nil
}

// ReadStaticSamlResponseFile reads SAML Response from a file.
func (c *Client) ReadStaticSamlResponseFile() error {
	assertions := c.Runtime.Saml.Assertions
	fp := assertions.GetPath()
	fh, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer fh.Close()
	c.Runtime.Saml.Assertions.Raw, err = ioutil.ReadAll(fh)
	if err != nil {
		return err
	}
	c.Runtime.Saml.Assertions.Plain = string(c.Runtime.Saml.Assertions.Raw[:])
	if strings.Contains(c.Runtime.Saml.Assertions.Plain, "\"SAMLResponse\"") {
		// This is the HTML input element containing SAML response
		i := strings.LastIndex(c.Runtime.Saml.Assertions.Plain, "value=\"")
		if i < 1 {
			return fmt.Errorf("Detected SAMLResponse, but value key not found")
		}
		c.Runtime.Saml.Assertions.Plain = c.Runtime.Saml.Assertions.Plain[i+7:]
		if i = strings.Index(c.Runtime.Saml.Assertions.Plain, "\""); i > 1 {
			c.Runtime.Saml.Assertions.Plain = c.Runtime.Saml.Assertions.Plain[:i]
		}
	}
	if !strings.Contains(c.Runtime.Saml.Assertions.Plain, "SAML:2.0:protocol") {
		// This is like base64 encoded SAML response
		c.Runtime.Saml.Assertions.Raw, err = base64.StdEncoding.DecodeString(c.Runtime.Saml.Assertions.Plain)
		if err != nil {
			return fmt.Errorf("No SAMLResponse and no SAML:2.0:protocol content")
		}
		c.Runtime.Saml.Assertions.Plain = string(c.Runtime.Saml.Assertions.Raw[:])
	}
	if !strings.Contains(c.Runtime.Saml.Assertions.Plain, "SAML:2.0:protocol") {
		return fmt.Errorf("SAML Response not found")
	}
	if err := xml.Unmarshal(c.Runtime.Saml.Assertions.Raw, &c.Runtime.Saml.Response); err != nil {
		return err
	}
	if c.Runtime.Saml.Response.Assertion.AttributeStatement == nil {
		return fmt.Errorf("SAML Response does not contain attribute statements: %v", c.Runtime.Saml.Response.Assertion)
	}
	c.Runtime.Saml.Attributes, err = c.Runtime.Saml.Response.GetAttributes()
	if err != nil {
		return fmt.Errorf("Failed to get attributes from SAML Response: %s", err)
	}
	return nil
}

func (c *Client) SetLogLevel(level log.Level) {
	log.SetLevel(level)
}

// InteractiveConfig propmts users for configuration data interactively.
func (c *Client) InteractiveConfig(s string) error {
	switch k := s; k {
	case "azure_tenant_id":
		if c.Config.Azure.TenantID != "" {
			return nil
		}
		fmt.Print("Enter Azure Tenant ID: ")
	case "azure_application_id":
		if c.Config.Azure.ApplicationID != "" {
			return nil
		}
		fmt.Print("Enter Azure Application ID for AWS Application: ")
	case "adfs_hostname":
		if c.Config.Adfs.Hostname != "" {
			return nil
		}
		fmt.Print("Enter ADFS Instance Hostname: ")
	case "email":
		if c.Config.Username != "" {
			return nil
		}
		fmt.Print("Enter email (or username): ")
	case "password":
		if c.Config.Password != "" {
			return nil
		}
		fmt.Printf("Enter password for %s: ", c.Config.Username)
	case "static_saml_response_file":
		if c.Config.Static.SamlResponseFile != "" {
			return nil
		}
		fmt.Print("Enter the path to the file containing SAML Response Claims: ")
	default:
		return fmt.Errorf("unsupported config item: %s", s)
	}
	timer := time.AfterFunc(time.Minute, func() {
		fmt.Fprintf(os.Stderr, "\nTimed out ...\n")
		os.Exit(1)
	})
	defer timer.Stop()

	reader := bufio.NewReader(os.Stdin)
	var v string
	var err error

	if s == "password" {
		p, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("Erred when processing password input: %s", err)
		}
		v = string(p)
		fmt.Fprintf(os.Stdout, "\n")
	} else {
		v, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("Erred when processing user input: %s", err)
		}
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return fmt.Errorf("No user input")
	}
	switch k := s; k {
	case "azure_tenant_id":
		return c.SetAzureTenantID(v)
	case "azure_application_id":
		return c.SetAzureApplicationID(v)
	case "adfs_hostname":
		return c.SetAdfsHostname(v)
	case "email":
		return c.SetUsername(v)
	case "password":
		return c.SetPassword(v)
	case "static_saml_response_file":
		return c.SetStaticSamlResponseFile(v)
	}
	return nil
}

// GetAuthenticationURL build ADFS Authentication URL.
func (c *Client) GetAuthenticationURL() error {
	if c.Runtime.AuthenticationURL != "" {
		return nil
	}
	if c.Config.Adfs.Hostname != "" {
		c.Runtime.AuthenticationURL = "https://" + c.Config.Adfs.Hostname +
			"/adfs/ls/IdpInitiatedSignOn.aspx?loginToRp=urn:amazon:webservices"
	}

	if c.Config.Azure.TenantID != "" {
		c.Runtime.AuthenticationURL = "https://login.microsoftonline.com/" +
			c.Config.Azure.TenantID + "/saml2"
	}
	return nil
}

// GetAdfsAuthenticationRequestBody build ADFS authentication request body.
func (c *Client) GetAdfsAuthenticationRequestBody() (url.Values, error) {
	v := url.Values{}
	if c.Config.Username == "" {
		return v, fmt.Errorf("No username found for ADFS authentication")
	}
	if c.Config.Password == "" {
		return v, fmt.Errorf("No password found for ADFS authentication")
	}
	v.Set("UserName", c.Config.Username)
	v.Set("Password", c.Config.Password)
	v.Set("AuthMethod", "FormsAuthentication")
	return v, nil
}

// GetAwsCredentials makes SAML request, authenticates to SAML IdP endpoint
// and receives SAML assertions back. Then, it sends the assertions to AWS STS
// service. The service responds with temporary credentials.
func (c *Client) GetAwsCredentials() ([]*AwsCredentials, error) {
	if err := c.GetAdfsMetadata(); err != nil {
		return nil, err
	}
	if err := c.GetSamlAssertions(); err != nil {
		return nil, err
	}
	if err := c.AssumeRoleWithSaml(); err != nil {
		return nil, err
	}
	return c.Aws.Credentials, nil
}
