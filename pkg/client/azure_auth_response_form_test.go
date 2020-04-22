package client

import (
	"io/ioutil"
	"path"
	"testing"
)

func TestParseAzureAuthResponseForm(t *testing.T) {
	testFailed := 0
	assetDir := "../../assets/tests"
	for i, test := range []struct {
		input      string
		exp        *AzureAuthResponseForm
		shouldFail bool
		shouldErr  bool
	}{
		{
			input: "azure.auth.response.form.html",
			exp: &AzureAuthResponseForm{
				Host: "signin.aws.amazon.com",
				Fields: map[string]string{
					"SAMLResponse": "TheTextOfSamlResponse",
				},
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
		resp, err := NewAzureAuthResponseFormFromBytes(content)
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

		if (len(resp.Fields) != len(test.exp.Fields)) && !test.shouldFail {
			t.Logf("FAIL: Test %d: input '%s', expected to pass, but failed due to mismatch %v (expected) vs %v (file)",
				i, test.input, test.exp.Fields, resp.Fields)
			testFailed++
			continue
		}

		if test.shouldFail {
			t.Logf("PASS: Test %d: input '%s', expected to fail, failed", i, test.input)
		} else {
			t.Logf("PASS: Test %d: input '%s', expected to pass, passed", i, test.input)
			t.Logf("  URL: %s", resp.URL)
			t.Logf("  Number of form fields: %d", len(resp.Fields))
			t.Logf("  Field SAMLResponse => %v", resp.Fields["SAMLResponse"])
		}
	}
	if testFailed > 0 {
		t.Fatalf("Failed %d tests", testFailed)
	}
}
