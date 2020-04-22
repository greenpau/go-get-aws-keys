package client

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	testFailed := 0
	for i, test := range []struct {
		awsRole      string
		awsAccountID string
		shouldFail   bool
	}{
		{
			awsRole:      "Administrator",
			awsAccountID: "MyAccount",
			shouldFail:   false,
		},
	} {
		cli := New()
		if err := cli.SetAwsRole(test.awsRole); err != nil {
			t.Logf("FAIL: Test %d (SetAssumeRole): expected to pass, but failed with: %v", i, err)
			testFailed++
			continue
		}
		if err := cli.SetAwsAccountID(test.awsAccountID); err != nil {
			t.Logf("FAIL: Test %d (SetAccount): expected to pass, but failed with: %v", i, err)
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
