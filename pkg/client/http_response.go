package client

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

type WebResponse struct {
	Success     bool
	Redirect    bool
	RedirectURL string
}

func parseWebResponse(resp *http.Response, body string) (*WebResponse, error) {
	r := &WebResponse{}
	log.Debugf("Response Body: %s", body)
	log.Debugf("Response HTTP Status Code: %d", resp.StatusCode)
	log.Debugf("Response HTTP Status: %s", resp.Status)
	return r, nil
}
