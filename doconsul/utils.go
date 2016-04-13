package doconsul

import (
	"fmt"
	"strings"

	"github.com/mitchellh/packer/packer"
)

func getImageIDfromDOArtifact(a packer.Artifact) (string, error) {
	stringArray := strings.Split(a.Id(), ":")
	if len(stringArray) != 2 {
		return "", fmt.Errorf("Error: imageID has invalid format: %s", a.Id())
	}
	return stringArray[len(stringArray)-1], nil
}
