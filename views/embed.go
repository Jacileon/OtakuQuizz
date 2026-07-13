package views

import "embed"

//go:embed layouts/*.html pages/*.html partials/*.html
var TemplatesFS embed.FS
