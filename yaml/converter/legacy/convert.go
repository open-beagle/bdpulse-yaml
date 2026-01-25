package legacy

import yaml "github.com/open-beagle/go-yaml/yaml/converter/legacy/internal"

// Convert converts the yaml configuration file from
// the legacy format to the 1.0+ format.
func Convert(d []byte, remote string) ([]byte, error) {
	return yaml.Convert(d, remote)
}
