package tulip

import _ "embed"

// BaseManifestYAML is the embedded base tulip manifest.
// This is baked into the binary so the CLI works without
// the source repository being present on disk.
//
//go:embed distributions/tulip/manifest.yaml
var BaseManifestYAML []byte
