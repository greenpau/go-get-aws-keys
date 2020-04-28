package main

import (
	"flag"
	"fmt"
	"github.com/greenpau/go-get-aws-keys/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var configFile string
	var azureTenantID, azureApplicationID string
	var adfsHostname string
	var staticSamlResponse string
	var emailAddress, password string
	var awsAccountID, awsRole, awsRegion, awsProfileName string
	var logLevel string
	var isShowVersion bool
	var isNoPrompt bool
	var outputCredFilePath string
	var outputEnvVarFilePath string
	cli := client.New()
	flag.StringVar(&configFile, "conf-file-name", "", "Path to configuration file")
	flag.StringVar(&emailAddress, "email", "", "Set email (or username) for authentication")
	flag.StringVar(&password, "password", "", "Set password for authentication")
	flag.StringVar(&azureTenantID, "adfs-azure-tenant-id", "", "Set Azure Tenant ID for ADFS authentication")
	flag.StringVar(&azureApplicationID, "adfs-azure-application-id", "", "Set Azure AWS Application ID for ADFS authentication")
	flag.StringVar(&adfsHostname, "adfs-enterprise-hostname", "", "Set hostname for enterprise ADFS authentication")
	flag.StringVar(&staticSamlResponse, "static-saml-file", "", "sets the path to the file with SAML Response claims")
	flag.StringVar(&awsAccountID, "aws-account-id", "", "AWS account ID")
	flag.StringVar(&awsRole, "aws-iam-role", "", "The name of AWS IAM Role")
	flag.StringVar(&awsRegion, "aws-region", "us-east-1", "AWS Region")
	flag.StringVar(&awsProfileName, "aws-profile-name", "default", "AWS Profile Name")
	flag.StringVar(&outputCredFilePath, "output-credentials-file", "~/.aws/credentials", "The path to write AWS credentials to")
	flag.StringVar(&outputEnvVarFilePath, "output-env-file", "~/.aws/environment", "The path to write AWS environment variables to")
	flag.BoolVar(&isNoPrompt, "no-prompt", false, "Disables prompting a user for required information")
	flag.StringVar(&logLevel, "log-level", "info", "logging severity level")
	flag.BoolVar(&isShowVersion, "version", false, "version information")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n%s - %s\n\n", cli.Info.Name, cli.Info.Description)
		fmt.Fprintf(os.Stderr, "Usage: %s [arguments]\n\n", cli.Info.Name)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nDocumentation: %s\n\n", cli.Info.Documentation)
	}
	flag.Parse()
	if isShowVersion {
		fmt.Fprintf(os.Stdout, "%s\n", cli.GetVersionInfo())
		os.Exit(0)
	}
	if level, err := log.ParseLevel(logLevel); err == nil {
		log.SetLevel(level)
		cli.SetLogLevel(level)
	} else {
		log.Fatalf(err.Error())
	}

	if configFile == "" {
		configFile = "go-get-aws-keys-config.yaml"
	}

	configName := strings.TrimSuffix(configFile, filepath.Ext(configFile))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigName(configName)
	viper.BindEnv("azure.tenant_id", "GGK_AZURE_TENANT_ID")
	viper.BindEnv("azure.application_id", "GGK_AZURE_AWS_APP_ID")
	viper.BindEnv("email", "GGK_EMAIL")
	viper.BindEnv("password", "GGK_PASSWORD")
	viper.AddConfigPath("$HOME/.aws")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetConfigType("yml")
	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("Error reading configuration file, %s", err)
	}
	if err := viper.Unmarshal(&cli.Config); err != nil {
		log.Fatalf("Error parsing configuration file: %s", err)
	}
	if v := viper.Get("azure.tenant_id"); v != nil {
		azureTenantID = v.(string)
	}
	if v := viper.Get("azure.application_id"); v != nil {
		azureApplicationID = v.(string)
	}
	if v := viper.Get("adfs.hostname"); v != nil {
		adfsHostname = v.(string)
	}
	if v := viper.Get("static.saml_response_file"); v != nil {
		staticSamlResponse = v.(string)
	}

	if emailAddress == "" {
		if v := viper.Get("email"); v != nil {
			emailAddress = v.(string)
		}
	}
	if password == "" {
		if v := viper.Get("password"); v != nil {
			password = v.(string)
		}
	}

	promptUser := []string{}

	if awsAccountID != "" && awsRole != "" {
		// user provided account name and the role via cli
		if cli.Config.Aws.Roles != nil {
			cli.Config.Aws.Roles = nil
		}
		role := map[string]string{
			"account_id":   awsAccountID,
			"name":         awsRole,
			"region":       awsRegion,
			"profile_name": awsProfileName,
		}
		if err := cli.RequestAwsRole(role); err != nil {
			log.Fatal(err)
		}
	} else {
		// user provided the roles via config file
		if v := viper.Get("aws.roles"); v != nil {
			for i, roles := range v.([]interface{}) {
				for k, v := range roles.(map[interface{}]interface{}) {
					switch k.(string) {
					case "account_id":
						cli.Config.Aws.Roles[i].AccountID = v.(string)
					case "role":
						cli.Config.Aws.Roles[i].Name = v.(string)
					case "region":
						cli.Config.Aws.Roles[i].DefaultRegion = v.(string)
					case "profile_name":
						cli.Config.Aws.Roles[i].ProfileName = v.(string)
					}
				}
			}
		}
		if err := cli.UpdateAwsRoles(); err != nil {
			log.Fatal(err)
		}
	}

	/* Populate configuration */
	enabledFeatures := []string{}
	if staticSamlResponse != "" {
		enabledFeatures = append(enabledFeatures, "static")
		if err := cli.SetStaticSamlResponseFile(staticSamlResponse); err != nil {
			log.Fatal(err)
		}
	}
	if azureTenantID != "" {
		enabledFeatures = append(enabledFeatures, "azure")
		if err := cli.SetAzureTenantID(azureTenantID); err != nil {
			log.Fatal(err)
		}
	}
	if adfsHostname != "" {
		enabledFeatures = append(enabledFeatures, "adfs")
		if err := cli.SetAdfsHostname(adfsHostname); err != nil {
			log.Fatal(err)
		}
	}
	if len(enabledFeatures) == 0 {
		log.Fatalf("must provide at least one way of obtaining SAML claims")
	}
	if len(enabledFeatures) != 1 {
		log.Fatalf("must provide only one way of obtaining SAML claims, provided %v", enabledFeatures)
	}

	if azureTenantID != "" {
		if azureApplicationID != "" {
			if err := cli.SetAzureApplicationID(azureApplicationID); err != nil {
				log.Fatal(err)
			}
		} else {
			promptUser = append(promptUser, "azure_application_id")
		}
	}

	if staticSamlResponse == "" {
		if emailAddress != "" {
			if err := cli.SetUsername(emailAddress); err != nil {
				log.Fatal(err)
			}
		} else {
			promptUser = append(promptUser, "email")
		}

		if password != "" {
			if err := cli.SetPassword(password); err != nil {
				log.Fatal(err)
			}
		} else {
			promptUser = append(promptUser, "password")
		}
	}

	if !isNoPrompt {
		for _, p := range promptUser {
			if err := cli.InteractiveConfig(p); err != nil {
				log.Fatalf("%s: failed to interactively prompt a user for %s: %s", cli.Info.Name, p, err)
			}
		}
	}

	if v := viper.ConfigFileUsed(); v != "" {
		cli.SetConfigFile(v)
	}

	awsCredentials, err := cli.GetAwsCredentials()
	if err != nil {
		log.Fatal(err)
	}

	for i, awsCredential := range awsCredentials {
		log.Debugf("AWS Access Keys #%d: %v", i, awsCredential)
		if err := awsCredential.WriteCredentialsFile(outputCredFilePath); err != nil {
			log.Fatal(err)
		}
	}
}
