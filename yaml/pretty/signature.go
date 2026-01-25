package pretty

import (
	"github.com/open-beagle/go-yaml/yaml"
)

// helper function pretty prints the signature resource.
func printSignature(w writer, v *yaml.Signature) {
	w.WriteString("---")
	w.WriteTagValue("version", v.Version)
	w.WriteTagValue("kind", v.Kind)
	w.WriteTagValue("hmac", v.Hmac)
	w.WriteByte('\n')
	w.WriteByte('\n')
}
