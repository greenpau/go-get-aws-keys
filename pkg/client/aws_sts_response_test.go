package client

import (
	"io/ioutil"
	"path"
	"testing"
)

func TestParseAwsStsResponse(t *testing.T) {
	testFailed := 0
	assetDir := "../../assets/tests"
	for i, test := range []struct {
		input      string
		exp        *AwsStsResponse
		byteSize   int
		shouldFail bool
		shouldErr  bool
	}{
		{
			input:      "aws.sts.response.1.json",
			exp:        &AwsStsResponse{},
			byteSize:   40,
			shouldFail: false,
			shouldErr:  false,
		},
	} {
		fp := path.Join(assetDir, test.input)
		content, err := ioutil.ReadFile(fp)
		if err != nil {
			t.Logf("FAIL: Test %d: failed reading '%s', error: %v", i, fp, err)
			testFailed++
			continue
		}
		resp, err := NewAwsStsResponseFromBytes(content)
		if err != nil {
			if !test.shouldErr {
				t.Logf("FAIL: Test %d: input '%s', expected to pass, but threw error: %v", i, test.input, err)
				testFailed++
				continue
			}
		} else {
			if test.shouldErr {
				t.Logf("FAIL: Test %d: input '%s', expected to throw error, but passed: %v", i, test.input, resp)
				testFailed++
				continue
			}
		}

		if (len(resp.Credentials.SecretAccessKey) != test.byteSize) && !test.shouldFail {
			t.Logf("FAIL: Test %d: input '%s', expected to pass, but failed due to byteSize of Credentials.SecretAccessKey [%d] != %d",
				i, test.input, len(resp.Credentials.SecretAccessKey), test.byteSize)
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
