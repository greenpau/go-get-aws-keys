package client

import (
	"io/ioutil"
	"path"
	"testing"
)

func TestParseAdfsAuthResponseForm(t *testing.T) {
	testFailed := 0
	assetDir := "../../assets/tests"
	for i, test := range []struct {
		input      string
		exp        *AdfsAuthResponseForm
		shouldFail bool
		shouldErr  bool
	}{
		{
			input: "adfs.auth.response.form.html",
			exp: &AdfsAuthResponseForm{
				Host: "login.microsoftonline.com",
				Fields: map[string]string{
					"wa":      "wsignin1.0",
					"wctx":    "estsredirect=2&amp;estsrequest=rQfd066002b7aa4882b08367a65ca0fd7b",
					"wresult": `<t:RequestSecurityTokenResponse xmlns:t="http://schemas.xmlsoap.org/ws/2005/02/trust"><t:KeyType>http://schemas.xmlsoap.org/ws/2005/05/identity/NoProofKey</t:KeyType></t:RequestSecurityTokenResponse>`,
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
		resp, err := NewAdfsAuthResponseFormFromBytes(content)
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
			t.Logf("  Field wresult => %v", resp.Fields["wresult"])
		}
	}
	if testFailed > 0 {
		t.Fatalf("Failed %d tests", testFailed)
	}
}
