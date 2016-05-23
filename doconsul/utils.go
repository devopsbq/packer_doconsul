package doconsul

import (
	"fmt"
	"log"
	"strings"

	"github.com/goware/urlx"
	"github.com/mitchellh/packer/packer"
)

const (
	defaultConsulHost    = "127.0.0.1"
	defaultConsulAPIPort = "8500"
)

func (c *Config) errorHandler(e []error) error {
	for _, err := range e {
		if err != nil {
			if c.IgnoreConnectionErrors {
				log.Printf("Ignored error: %s", err.Error())
				continue
			}
			return err
		}
	}
	return nil
}

func getImageIDfromDOArtifact(a packer.Artifact) (map[string]string, error) {
	stringArray := strings.Split(a.Id(), ":")
	if len(stringArray) != 2 {
		return nil, fmt.Errorf("Error: imageID has invalid format: %s", a.Id())
	}
	return map[string]string{stringArray[0]: stringArray[1]}, nil
}

func getImageIDfromAWSArtifact(a packer.Artifact) (map[string]string, error) {
	stringArray := strings.Split(a.Id(), ",")
	regionAndAMI := make(map[string]string)
	for _, value := range stringArray {
		split := strings.Split(value, ":")
		if len(split) != 2 {
			return nil, fmt.Errorf("Error: imageID has invalid format: %s", a.Id())
		}
		regionAndAMI[split[0]] = split[1]
	}
	return regionAndAMI, nil
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
