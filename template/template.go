package template

import "embed"

//go:embed *.go.tmpl
var FS embed.FS

func ReadTemplate(name string) ([]byte, error) {
	content, err := FS.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return content, nil
}
