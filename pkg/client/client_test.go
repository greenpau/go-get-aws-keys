package client

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	testFailed := 0
	for i, test := range []struct {
		awsRole        string
		awsAccountID   string
		awsRegion      string
		awsProfileName string
		shouldFail     bool
	}{
		{
			awsRole:        "Administrator",
			awsAccountID:   "MyAccount",
			awsRegion:      "us-east-1",
			awsProfileName: "default",
			shouldFail:     false,
		},
	} {
		cli := New()
		role := map[string]string{
			"account_id":   test.awsAccountID,
			"name":         test.awsRole,
			"region":       test.awsRegion,
			"profile_name": test.awsProfileName,
		}
		if err := cli.RequestAwsRole(role); err != nil {
			t.Logf("FAIL: Test %d (RequestAwsRole): expected to pass, but failed with: %v", i, err)
			testFailed++
			continue
		}
		if test.shouldFail {
			t.Logf("FAIL: Test %d: expected to fail, but passed", i)
			testFailed++
			continue
		}
		t.Logf("PASS: Test %d: expected to pass: passed", i)
	}
	if testFailed > 0 {
		t.Fatalf("Failed %d tests", testFailed)
	}
}
