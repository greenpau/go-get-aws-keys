package client

import (
	"io/ioutil"
	"path"
	"testing"
)

func TestParseAdfsAuthForm(t *testing.T) {
	testFailed := 0
	assetDir := "../../assets/tests"
	for i, test := range []struct {
		input      string
		exp        *AdfsAuthForm
		shouldFail bool
		shouldErr  bool
	}{
		{
			input: "adfs.auth.form.html",
			exp: &AdfsAuthForm{
				Host: "adfs.contoso.com",
			},
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
		resp, err := NewAdfsAuthFormFromBytes(content)
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

		if (resp.Host != test.exp.Host) && !test.shouldFail {
			t.Logf("FAIL: Test %d: input '%s', expected to pass, but failed due to mismatch %s (expected) vs %s (file)",
				i, test.input, test.exp.Host, resp.Host)
			testFailed++
			continue
		}

		if test.shouldFail {
			t.Logf("PASS: Test %d: input '%s', expected to fail, failed", i, test.input)
		} else {
			t.Logf("PASS: Test %d: input '%s', expected to pass, passed", i, test.input)
			t.Logf("  URL: %s", resp.URL)
		}
	}
	if testFailed > 0 {
		t.Fatalf("Failed %d tests", testFailed)
	}
}
