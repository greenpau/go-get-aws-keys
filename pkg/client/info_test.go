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
				Name:    "go-get-aws-keys",
				Version: "1.0.0",
				Git: GitInfo{
					Branch: "master",
					Commit: "63dc8b9",
				},
				Build: BuildInfo{
					OperatingSystem: "linux",
					Architecture:    "amd64",
					User:            "greenpau",
					Date:            "2019-09-18",
				},
			},
			byteSize:   115,
			shouldFail: false,
		},
		{
			input: Info{
				Name:    "go-get-aws-keys",
				Version: "1.0.0",
				Git: GitInfo{
					Branch: "master",
					Commit: "63dc8b9",
				},
			},
			byteSize:   66,
			shouldFail: false,
		},
		{
			input: Info{
				Name:    "go-get-aws-keys",
				Version: "1.0.0",
				Build: BuildInfo{
					OperatingSystem: "linux",
					Architecture:    "amd64",
					User:            "greenpau",
					Date:            "2019-09-18",
				},
			},
			byteSize:   82,
			shouldFail: false,
		},

		{
			input: Info{
				Name: "go-get-aws-keys",
			},
			byteSize:   28,
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
