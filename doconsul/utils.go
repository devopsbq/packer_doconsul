package doconsul

import (
	"fmt"
	"strings"

	"github.com/goware/urlx"
	"github.com/mitchellh/packer/packer"
)

const (
	defaultConsulHost    = "127.0.0.1"
	defaultConsulAPIPort = "8500"
)

func (c *Config) errorHandler(e error) error {
	if c.IgnoreConnectionErrors {
		return nil
	}
	return e
}

func getImageIDfromDOArtifact(a packer.Artifact) (string, error) {
	stringArray := strings.Split(a.Id(), ":")
	if len(stringArray) != 2 {
		return "", fmt.Errorf("Error: imageID has invalid format: %s", a.Id())
	}
	return stringArray[len(stringArray)-1], nil
}

func parseConsulAddress(address string) (string, error) {
	consulURL, err := urlx.Parse(address)
	if err != nil {
		return "", err
	}
	consulHost, consulPort, err := urlx.SplitHostPort(consulURL)
	if err != nil {
		return "", err
	}
	if consulPort == "" {
		consulPort = defaultConsulAPIPort
	}
	if consulHost == "" {
		consulHost = defaultConsulHost
	}

	return fmt.Sprintf("%s:%s", consulHost, consulPort), nil
}
