package client

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

// GetAdfsMetadata fetches the matadata about an inssuer.
func (c *Client) GetAdfsMetadata() error {
	if !c.IsMetadataNeeded() {
		return nil
	}
	if c.IsMetadataExists() {
		log.Debugf("Metadata file exists: %s", c.Runtime.Metadata.File.Path)
		if err := c.ReadMetadataFromFile(); err != nil {
			return fmt.Errorf("Error reading metadata from %s: %s", c.Runtime.Metadata.File.Path, err)
		}
		return nil
	}
	resp, err := http.Get(c.Runtime.Metadata.URL)
	if err != nil {
		return fmt.Errorf("Error querying metadata @ %s: %s", c.Runtime.Metadata.URL, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response data from %s: %s", c.Runtime.Metadata.URL, err)
	}
	c.Runtime.Metadata.Raw = body
	c.Runtime.Metadata.Plain = string(c.Runtime.Metadata.Raw[:])
	if err := c.WriteMetadataToFile(); err != nil {
		return fmt.Errorf("Error writing metadata to %s: %s", c.Runtime.Metadata.File.Path, err)
	}
	return nil
}
