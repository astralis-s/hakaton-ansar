// Package openapi embeds the generated OpenAPI specification for the PUBLIC API.
// In later phases `make swag` regenerates swagger.json/swagger.yaml here
// (Go output is disabled via --outputTypes json,yaml so this package stays the
// only .go file in the directory).
package openapi

import _ "embed"

//go:embed swagger.json
var SpecJSON []byte
