package build

import (
	_ "embed"
)

//go:embed cert.pem
var Certificate []byte
