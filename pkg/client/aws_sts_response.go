package client

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// AwsStsResponseCredentials contains the Credentials part of AwsStsResponse.
type AwsStsResponseCredentials struct {
	SecretAccessKey string
	SessionToken    string
	//Expiration      string
	//Expiration  time.Time
	AccessKeyId string
}

// AwsStsResponseAssumedRoleUser contains the AssumedRoleUser part of AwsStsResponse.
type AssumedRoleUser struct {
	AssumedRoleId string
	Arn           string
}

// AwsStsResponse contains successful AWS STS service response.
type AwsStsResponse struct {
	SubjectType     string
	AssumedRoleUser *AssumedRoleUser
	Audience        string
	NameQualifier   string
	Credentials     *AwsStsResponseCredentials
	Subject         string
	Issuer          string
}

// AwsStsResponseMetadata is the metadata associated with
// HTTP POST to AWS STS API endpoint.
type AwsStsResponseMetadata struct {
	RequestId string
}

// AwsStsAssumeRoleWithSAMLResult is the result of the
// HTTP POST to AWS STS API endpoint.
type AwsStsAssumeRoleWithSAMLResult struct {
	AssumeRoleWithSAMLResult *AwsStsResponse
}

// AwsStsAssumeRoleWithSAMLResponse is the response to HTTP POST to
// AWS STS APIendpoint.
type AwsStsAssumeRoleWithSAMLResponse struct {
	AssumeRoleWithSAMLResult *AwsStsResponse
	ResponseMetadata         AwsStsResponseMetadata
}

type AwsStsResponseBody struct {
	AssumeRoleWithSAMLResponse AwsStsAssumeRoleWithSAMLResponse
}

type AwsStsResponseError struct {
	Code    string
	Message string
	Type    string
}

type AwsStsErrorResponseBody struct {
	Error AwsStsResponseError
	AwsStsResponseMetadata
}

// NewAwsStsResponseFromString returns AwsStsResponse instance from an input string.
func NewAwsStsResponseFromString(s string) (*AwsStsResponse, error) {
	return NewAwsStsResponseFromBytes([]byte(s))
}

// NewAwsStsResponseFromBytes returns AwsStsResponse instance from an input byte array.
func NewAwsStsResponseFromBytes(s []byte) (*AwsStsResponse, error) {
	if bytes.Contains(s, []byte("\"Error\":")) {
		errResp := &AwsStsErrorResponseBody{}
		if err := json.Unmarshal(s, errResp); err != nil {
			return nil, fmt.Errorf("parsing error: %s, AWS STS response: %s", err, string(s[:]))
		}
		return nil, fmt.Errorf("AWS STS %s: %s: %s ", errResp.RequestId, errResp.Error.Code, errResp.Error.Message)
	}
	if bytes.Contains(s, []byte("ResponseMetadata")) {
		extResp := &AwsStsResponseBody{}
		if err := json.Unmarshal(s, extResp); err != nil {
			return nil, fmt.Errorf("parsing error: %s, AWS STS response: %s", err, string(s[:]))
		}
		resp := extResp.AssumeRoleWithSAMLResponse.AssumeRoleWithSAMLResult
		return resp, nil
	}

	resp := &AwsStsResponse{}
	if err := json.Unmarshal(s, resp); err != nil {
		return nil, fmt.Errorf("parsing error: %s, AWS STS response: %s", err, string(s[:]))
	}
	return resp, nil
}
