package client

import (
	"io/ioutil"
	"path"
	"testing"
)

func TestWriteAwsCredentials(t *testing.T) {
	testFailed := 0
	assetDir := "../../assets/tests"
	for i, test := range []struct {
		input      string
		shouldFail bool
		shouldErr  bool
	}{
		{
			input:      "aws.sts.response.1",
			shouldFail: false,
			shouldErr:  false,
		},
	} {
		stsFilePath := path.Join(assetDir, test.input+".json")
		content, err := ioutil.ReadFile(stsFilePath)
		if err != nil {
			t.Logf("FAIL: Test %d: failed reading '%s', error: %v", i, stsFilePath, err)
			testFailed++
			continue
		}

		awsStsResponse, _ := NewAwsStsResponseFromBytes(content)
		awsCredentials, _ := NewAwsCredentialsFromStsResponse(awsStsResponse)

		if awsCredentials == nil {
			if !test.shouldFail {
				t.Logf("FAIL: Test %d: input '%s', expected to pass, but failed due to awsCredentials being nil", i, test.input)
				testFailed++
			}
			t.Logf("PASS: Test %d: input '%s', expected to pass, passed", i, test.input)
			continue
		}

		envFilePath := path.Join(assetDir, test.input+".env")
		if err := awsCredentials.WriteEnvVarsFile(envFilePath); err != nil {
			if !test.shouldFail {
				t.Logf("FAIL: Test %d: input '%s', expected to pass, but failed: %s",
					i, test.input, err)
				testFailed++
			}
		}

		iniFilePath := path.Join(assetDir, test.input+".ini")
		if err := awsCredentials.WriteCredentialsFile(iniFilePath, "default"); err != nil {
			if !test.shouldFail {
				t.Logf("FAIL: Test %d: input '%s', expected to pass, but failed: %s",
					i, test.input, err)
				testFailed++
			}
		}

		if test.shouldFail {
			t.Logf("PASS: Test %d: input '%s', expected to fail, failed", i, test.input)
		} else {
			t.Logf("PASS: Test %d: input '%s', expected to pass, passed", i, test.input)
		}
	}
	if testFailed > 0 {
		t.Fatalf("Failed %d tests", testFailed)
	}
}

func TestIsValidAwsCredentials(t *testing.T) {
	testFailed := 0
	assetDir := "../../assets/tests"
	for i, test := range []struct {
		input      string
		creds      *AwsCredentials
		shouldFail bool
		shouldErr  bool
	}{
		{
			input: "aws.sts.response.1.json",
			creds: &AwsCredentials{
				AccessKeyId: "xxxx",
			},
			shouldFail: true,
			shouldErr:  false,
		},
		{
			input: "aws.sts.response.1.json",
			creds: &AwsCredentials{
				AccessKeyId: "ASBGQSJR7ZAFSXODTUMO",
			},
			shouldFail: false,
			shouldErr:  false,
		},
		{
			input: "aws.sts.response.2.json",
			creds: &AwsCredentials{
				AccessKeyId: "xxxx",
			},
			shouldFail: true,
			shouldErr:  true,
		},
	} {
		fp := path.Join(assetDir, test.input)
		content, err := ioutil.ReadFile(fp)
		if err != nil {
			t.Logf("FAIL: Test %d: failed reading '%s', error: %v", i, fp, err)
			testFailed++
			continue
		}
		awsStsResponse, _ := NewAwsStsResponseFromBytes(content)
		awsCredentials, err := NewAwsCredentialsFromStsResponse(awsStsResponse)
		if err != nil {
			if !test.shouldErr {
				t.Logf("FAIL: Test %d: input '%s', expected to pass, but threw error: %v", i, test.input, err)
				testFailed++
				continue
			}
		} else {
			if test.shouldErr {
				t.Logf("FAIL: Test %d: input '%s', expected to throw error, but passed: %v", i, test.input, awsCredentials)
				testFailed++
				continue
			}
		}

		if awsCredentials == nil {
			if !test.shouldFail {
				t.Logf("FAIL: Test %d: input '%s', expected to pass, but failed due to awsCredentials being nil", i, test.input)
				testFailed++
			}
			t.Logf("PASS: Test %d: input '%s', expected to pass, passed", i, test.input)
			continue
		}

		if (awsCredentials.AccessKeyId != test.creds.AccessKeyId) && !test.shouldFail {
			t.Logf("FAIL: Test %d: input '%s', expected to pass, but failed due to mismatch of AccessKeyId [%s] (expected) != %s",
				i, test.input, awsCredentials.AccessKeyId, test.creds.AccessKeyId)
			testFailed++
			continue
		}

		if test.shouldFail {
			t.Logf("PASS: Test %d: input '%s', expected to fail, failed", i, test.input)
		} else {
			t.Logf("PASS: Test %d: input '%s', expected to pass, passed", i, test.input)
		}
	}
	if testFailed > 0 {
		t.Fatalf("Failed %d tests", testFailed)
	}
}
