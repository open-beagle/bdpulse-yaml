//go:build oss
// +build oss

package converter

// Convert converts the yaml configuration file from
// the legacy format to the 1.0+ format.
func Convert(d []byte, m Metadata) ([]byte, error) {
	return d, nil
}

// ConvertString converts the yaml configuration file from
// the legacy format to the 1.0+ format.
func ConvertString(s string, m Metadata) (string, error) {
	return s, nil
}
