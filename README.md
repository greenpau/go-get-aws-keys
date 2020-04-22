# go-get-aws-keys

<a href="https://github.com/greenpau/go-get-aws-keys/actions/" target="_blank"><img src="https://github.com/greenpau/go-get-aws-keys/workflows/build/badge.svg?branch=master"></a>
<a href="https://pkg.go.dev/github.com/greenpau/go-get-aws-keys" target="_blank"><img src="https://img.shields.io/badge/godoc-reference-blue.svg"></a>

Obtain AWS STS Tokens by authenticating to ADFS via Azure, getting SAML
claim from Azure, and passing the claims to AWS STS service.

**Who is this tool for?**

* Your organization manages AWS role assignments in Azure's AWS application
* You use office.com portal to login to AWS
* You authenticate via ADFS

## Getting Started

### Building from Source

The tool is a single binary:
* Linux/MAC: `go-get-aws-keys`
* Windows: `go-get-aws-keys.exe`

Run the following commands:

```bash
git clone git@github.com:greenpau/go-get-aws-keys.git
cd go-get-aws-keys
make
make BUILD_OS="darwin"
make BUILD_OS="windows" BUILD_ARCH="amd64"
make BUILD_OS="windows" BUILD_ARCH="386"
```

The commands build binaries for various operating systems:

```bash
$ find bin/ -type f
bin/linux-amd64/go-get-aws-keys
bin/darwin-amd64/go-get-aws-keys
bin/windows-amd64/go-get-aws-keys.exe
bin/windows-386/go-get-aws-keys.exe
```

### Configuration File

The configuration file name is `go-get-aws-keys-config.yaml`.
The location of the file depends on the type of an operation
system it is used on (see instructions below).

```yaml
---
azure:
  tenant_id: '9c5399e3-e3e4-49aa-b6c7-e27d618ae206'
  application_id: 'f4cd2b32-6d0d-423d-85ce-9acc0318a4fe'
aws:
  roles:
  - account_id: '000000000001'
    role: 'Administrator'
    region: 'us-east-1'
    profile_name: 'default'
  - account_id: '000000000002'
    role: 'Administrator'
    region: 'us-east-1'
email: 'jsmith@contoso.com'
password: 'My@Password' # nor recommended
```

The configuration file has `azure` section for Azure specific configuration.
Please reach out to Azure AD administrator to provide you with
Azure Tenant ID and the ID for the AWS application in Azure.

The configuration file also has `aws` section for defining the
roles that a user want to assume.

After `go-get-aws-keys` gets temporary STS credentials it puts them
into `.aws/credentials` file. The value of the `profile_name` key in
`aws` section defines the name of the profile the tool will create or
update:
  - If the profile already exists in `.aws/credentials` file, then that
    specific section will be overwritten.
  - If the profile name does not exist, then it the profile will be
    appended to the credentials file.
  - If the profile name is not being set in the configuration file, then
    the profile name will match the following pattern
    `ggk-<Account ID>-<Role Name>`

This is what you would expect to see when invoking `go-get-aws-keys`:

```
$ go-get-aws-keys
Enter password for jsmith@contoso.com:
INFO[0003] Added default aws credentials profile to /home/jsmith/.aws/credentials
INFO[0003] Added ggk-000000000002-Administrator aws credentials profile to /home/jsmith/.aws/credentials
```

#### Linux and MAC OS

* Create a configuration file: `~/.aws/go-get-aws-keys-config.yaml`
* Place `go-get-aws-keys` in your executable path

#### Windows

* Go to "View Advanced System Settings" in "System Properties"
* Click "Environment Variables"
* Amend "Path" environment variable for your user by adding: `C:\Users\<username>\AppData\Local\Programs\go-get-aws-keys`
* Create `C:\Users\<username>\AppData\Local\Programs\go-get-aws-keys` directory
* Unpack `go-get-aws-keys.exe` in `go-get-aws-keys-1.0.0.windows-amd64.zip` to the above `go-get-aws-keys` directory
* Unpack `go-get-aws-keys-config.yaml` in `go-get-aws-keys-1.0.0.windows-amd64.zip` to `C:\Users\<username>\.aws` directory

Alternatively, a user may set up the environment variable in the following way:

```bash
setx path "%path%;%userprofile%\AppData\Local\Programs\go-get-aws-keys"
```
