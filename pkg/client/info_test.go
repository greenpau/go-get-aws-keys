package client

import (
	"testing"
)

func TestGetVersionInfo(t *testing.T) {
	testFailed := 0
	for i, test := range []struct {
		input      Info
		byteSize   int
		shouldFail bool
	}{
		{
			input: Info{
				Name: "go-get-aws-keys",
			},
			byteSize:   16,
			shouldFail: false,
		},
	} {
		cli := New()
		cli.Info = test.input
		resp := cli.GetVersionInfo()
		if (len(resp) != test.byteSize) && !test.shouldFail {
			t.Logf("FAIL: Test %d: input '%s', expected to pass, but failed due to byteSize of version info [%d] != %d",
				i, test.input, len(resp), test.byteSize)
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
