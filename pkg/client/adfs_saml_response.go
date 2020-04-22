package client

// This code a stripped down version of https://github.com/glucn/saml/blob/master/internal/saml/response.go.
// It does not include verification and signing, because they are not necessary for
// this package.

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	AwsRoleSessionNameAttribute = "https://aws.amazon.com/SAML/Attributes/RoleSessionName"
	AwsRoleAttribute            = "https://aws.amazon.com/SAML/Attributes/Role"
	AwsSessionDurationAttribute = "https://aws.amazon.com/SAML/Attributes/SessionDuration"
)

// SamlResponse is the structure holding SAMLv2 response.
type SamlResponse struct {
	XMLName      xml.Name           `xml:"urn:oasis:names:tc:SAML:2.0:protocol Response"`
	ID           string             `xml:"ID,attr"`
	Version      string             `xml:"Version,attr"`
	IssueInstant time.Time          `xml:"IssueInstant,attr"`
	Destination  string             `xml:"Destination,attr,omitempty"`
	Issuer       SamlProtocolIssuer `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
	Status       SamlProtocolStatus `xml:"urn:oasis:names:tc:SAML:2.0:protocol Status"`
	Assertion    SamlAssertion      `xml:"urn:oasis:names:tc:SAML:2.0:assertion Assertion"`
}

// SamlProtocolStatusCode is a structure holding the StatusCode of SAMLv2 response.
type SamlProtocolStatusCode struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol StatusCode"`
	Value   string   `xml:"Value,attr"`
}

// SamlProtocolStatus is a structure holding the Status SAMLv2 response.
type SamlProtocolStatus struct {
	XMLName    xml.Name               `xml:"urn:oasis:names:tc:SAML:2.0:protocol Status"`
	StatusCode SamlProtocolStatusCode `xml:"urn:oasis:names:tc:SAML:2.0:protocol StatusCode"`
}

// SamlProtocolIssuer is a structure holding the Issuer of SAMLv2 response.
type SamlProtocolIssuer struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
	Format  string   `xml:"Format,attr"`
	Issuer  string   `xml:",chardata"`
}

// SamlAssertion is a structure holding SAMLv2 response assertion.
type SamlAssertion struct {
	XMLName      xml.Name  `xml:"urn:oasis:names:tc:SAML:2.0:assertion Assertion"`
	ID           string    `xml:"ID,attr"`
	Version      string    `xml:"Version,attr"`
	IssueInstant time.Time `xml:"IssueInstant,attr"`

	Issuer             string                           `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
	Subject            SamlAssertionSubject             `xml:"urn:oasis:names:tc:SAML:2.0:assertion Subject"`
	Conditions         SamlAssertionConditions          `xml:"urn:oasis:names:tc:SAML:2.0:assertion Conditions"`
	AuthnStatement     SamlAssertionAuthnStatement      `xml:"urn:oasis:names:tc:SAML:2.0:assertion AuthnStatement"`
	AttributeStatement *SamlAssertionAttributeStatement `xml:"urn:oasis:names:tc:SAML:2.0:assertion AttributeStatement,omitempty"`
}

// SamlAssertionNameID is TBD.
type SamlAssertionNameID struct {
	XMLName         xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion NameID"`
	SPNameQualifier string   `xml:"SPNameQualifier,attr,omitempty"`
	Format          string   `xml:"Format,attr"`
	ID              string   `xml:",chardata"`
}

// SamlAssertionSubjectConfirmationData is TBD.
type SamlAssertionSubjectConfirmationData struct {
	XMLName      xml.Name  `xml:"urn:oasis:names:tc:SAML:2.0:assertion SubjectConfirmationData"`
	NotOnOrAfter time.Time `xml:"NotOnOrAfter,attr"`
	Recipient    string    `xml:"Recipient,attr"`
	InResponseTo string    `xml:"InResponseTo,attr,omitempty"`
}

// SamlAssertionSubjectConfirmation is TBD.
type SamlAssertionSubjectConfirmation struct {
	XMLName xml.Name                             `xml:"urn:oasis:names:tc:SAML:2.0:assertion SubjectConfirmation"`
	Method  string                               `xml:"Method,attr"`
	Data    SamlAssertionSubjectConfirmationData `xml:"urn:oasis:names:tc:SAML:2.0:assertion SubjectConfirmationData"`
}

// SamlAssertionSubject is TBD.
type SamlAssertionSubject struct {
	XMLName      xml.Name                         `xml:"urn:oasis:names:tc:SAML:2.0:assertion Subject"`
	NameID       SamlAssertionNameID              `xml:"urn:oasis:names:tc:SAML:2.0:assertion NameID"`
	Confirmation SamlAssertionSubjectConfirmation `xml:"urn:oasis:names:tc:SAML:2.0:assertion SubjectConfirmation"`
}

// SamlAssertionConditions is TBD.
type SamlAssertionConditions struct {
	XMLName             xml.Name                         `xml:"urn:oasis:names:tc:SAML:2.0:assertion Conditions"`
	NotBefore           time.Time                        `xml:"NotBefore,attr"`
	NotOnOrAfter        time.Time                        `xml:"NotOnOrAfter,attr"`
	AudienceRestriction SamlAssertionAudienceRestriction `xml:"urn:oasis:names:tc:SAML:2.0:assertion AudienceRestriction"`
}

// SamlAssertionAudienceRestriction is TBD.
type SamlAssertionAudienceRestriction struct {
	XMLName  xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion AudienceRestriction"`
	Audience string   `xml:"urn:oasis:names:tc:SAML:2.0:assertion Audience"`
}

// SamlAssertionAuthnStatement is TBD.
type SamlAssertionAuthnStatement struct {
	XMLName             xml.Name                  `xml:"urn:oasis:names:tc:SAML:2.0:assertion AuthnStatement"`
	AuthnInstant        time.Time                 `xml:"AuthnInstant,attr"`
	SessionNotOnOrAfter time.Time                 `xml:"SessionNotOnOrAfter,attr"`
	SessionIndex        string                    `xml:"SessionIndex,attr"`
	AuthnContext        SamlAssertionAuthnContext `xml:"urn:oasis:names:tc:SAML:2.0:assertion AuthnContext"`
}

// SamlAssertionAuthnContext is TBD.
type SamlAssertionAuthnContext struct {
	XMLName              xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion AuthnContext"`
	AuthnContextClassRef string   `xml:"urn:oasis:names:tc:SAML:2.0:assertion AuthnContextClassRef"`
}

// SamlAssertionAttributeStatement is TBD.
type SamlAssertionAttributeStatement struct {
	XMLName    xml.Name                 `xml:"urn:oasis:names:tc:SAML:2.0:assertion AttributeStatement"`
	Attributes []SamlAssertionAttribute `xml:"urn:oasis:names:tc:SAML:2.0:assertion Attribute"`
}

// SamlAssertionAttribute is TBD.
type SamlAssertionAttribute struct {
	XMLName    xml.Name                      `xml:"urn:oasis:names:tc:SAML:2.0:assertion Attribute"`
	Name       string                        `xml:"Name,attr"`
	NameFormat string                        `xml:"NameFormat,attr"`
	Values     []SamlAssertionAttributeValue `xml:"urn:oasis:names:tc:SAML:2.0:assertion AttributeValue"`
}

// SamlAssertionAttributeValue is TBD.
type SamlAssertionAttributeValue struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion AttributeValue"`
	XMLNS   string   `xml:"xmlns:xs,attr"`
	Type    string   `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	Value   string   `xml:",chardata"`
}

// SamlResponseData is TBD
type SamlResponseData struct {
	Aws struct {
		Roles                   []*AwsRole
		SessionName             string
		SessionDuration         int
		SessionEndTimestamp     time.Time
		SessionStartTimestamp   time.Time
		AuthenticateByTimestamp time.Time
	}
	Issuer  string
	Success bool
	Claims  []*SamlClaim
}

// SamlClaim is TBD.
type SamlClaim struct {
	Type  string
	Value string
}

// AwsRole is TBD.
type AwsRole struct {
	Raw                 string
	AccountID           string
	Name                string
	RoleARN             string
	IdentityProviderARN string
	ProfileName         string
	DefaultRegion       string
}

func ParseAwsRole(s string) (*AwsRole, error) {
	r := AwsRole{}
	r.Raw = s
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return &r, fmt.Errorf("the passed value is expected to have two ARNs, but it has %d", len(parts))
	}
	if !strings.HasPrefix(parts[1], "arn:aws:iam::") {
		return &r, fmt.Errorf("the passed Role ARN has no 'arn:aws:iam::' prefix, %s", parts[1])
	}
	r.IdentityProviderARN = parts[0]
	r.RoleARN = parts[1]
	roleParts := strings.Split(r.RoleARN, ":")
	if len(roleParts) != 6 {
		return &r, fmt.Errorf("the passed Role ARN does not match expected format %d parts: arn:aws:iam::<account_id>:role/<role_name>", len(roleParts))
	}
	r.AccountID = roleParts[4]
	roleNameParts := strings.Split(roleParts[5], "/")
	if len(roleNameParts) != 2 {
		return &r, fmt.Errorf("the passed Role name does not match expected format (%d parts): role/<role_name>", len(roleNameParts))
	}
	r.Name = roleNameParts[1]
	return &r, nil
}

func ParseSamlResponseClaim(t, s string) *SamlClaim {
	r := SamlClaim{}
	r.Type = t
	r.Value = s
	return &r

}

func (r *SamlResponse) GetAttributes() (*SamlResponseData, error) {
	resp := &SamlResponseData{}
	resp.Aws.Roles = []*AwsRole{}
	resp.Claims = []*SamlClaim{}
	if r.Assertion.AttributeStatement == nil {
		r.Assertion.AttributeStatement = &SamlAssertionAttributeStatement{}
	}
	for _, attr := range r.Assertion.AttributeStatement.Attributes {
		if attr.Name == "" {
			continue
		}
		if len(attr.Values) == 0 {
			continue
		}
		switch attr.Name {
		case AwsRoleSessionNameAttribute:
			resp.Aws.SessionName = attr.Values[0].Value
			resp.Claims = append(resp.Claims, ParseSamlResponseClaim(attr.Name, attr.Values[0].Value))
		case AwsSessionDurationAttribute:
			if i, err := strconv.Atoi(attr.Values[0].Value); err == nil {
				resp.Aws.SessionDuration = i
				resp.Claims = append(resp.Claims, ParseSamlResponseClaim(attr.Name, attr.Values[0].Value))
			}
		case AwsRoleAttribute:
			for _, entry := range attr.Values {
				if entry.Value == "" {
					continue
				}
				role, err := ParseAwsRole(entry.Value)
				if err != nil {
					return nil, fmt.Errorf("Failed to parse SAML Role, value: %s, error %s", entry.Value, err)
				}
				resp.Aws.Roles = append(resp.Aws.Roles, role)
				resp.Claims = append(resp.Claims, ParseSamlResponseClaim(attr.Name, attr.Values[0].Value))
			}
		default:
			resp.Claims = append(resp.Claims, ParseSamlResponseClaim(attr.Name, attr.Values[0].Value))
		}
	}
	if len(resp.Aws.Roles) == 0 {
		return nil, fmt.Errorf("AWS Roles not found: %s", AwsRoleAttribute)
	}
	if resp.Aws.SessionName == "" {
		return nil, fmt.Errorf("AWS Session Name not found: %s", AwsRoleSessionNameAttribute)
	}
	resp.Aws.AuthenticateByTimestamp = r.Assertion.Subject.Confirmation.Data.NotOnOrAfter
	resp.Aws.SessionStartTimestamp = r.Assertion.Conditions.NotBefore
	resp.Aws.SessionEndTimestamp = r.Assertion.Conditions.NotOnOrAfter
	resp.Issuer = r.Assertion.Issuer
	resp.Success = strings.Contains(r.Status.StatusCode.Value, "status:Success")
	return resp, nil
}
