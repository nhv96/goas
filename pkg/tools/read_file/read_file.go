package readfile

import (
	"fmt"
	"os"

	"github.com/nhv96/goas/pkg/tools"
)

// ReadFile is a tool to generate file with content
type ReadFile struct {
	tools.Base
}

// Init returns a tooler
func Init() tools.Tooler {
	t := ReadFile{
		Base: tools.Base{
			Type: "function",
			Function: tools.Function{
				Name:        "read_file",
				Description: "A tool used to read a file's content.",
				Parameters: tools.Parameters{
					Type:     "object",
					Required: []string{"path"},
					Properties: map[string]map[string]any{
						"path": {
							"type":        "string",
							"description": "The file path to be read, must be **absolute** path including file name and file extension",
						},
					},
				},
			},
		},
	}
	return t
}

// Run runs the tool with arguments
func (t ReadFile) Run(args any) (any, error) {
	in, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("wrong input")
	}

	path := in["path"].(string)

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("failed to read file %s with error: %s", path, err.Error()), nil
	}

	return string(content), nil
}
