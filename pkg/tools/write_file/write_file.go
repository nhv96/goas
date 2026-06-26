package writefile

import (
	"fmt"
	"os"

	"github.com/nhv96/goas/pkg/tools"
)

// WriteFile is a tool to generate file with content
type WriteFile struct {
	tools.Base
}

// Init returns a tooler
func Init() tools.Tooler {
	t := WriteFile{
		Base: tools.Base{
			Type: "function",
			Function: tools.Function{
				Name:        "write_file",
				Description: "A tool used to generate a file with its content.",
				Parameters: tools.Parameters{
					Type:     "object",
					Required: []string{"path", "content"},
					Properties: map[string]map[string]any{
						"path": {
							"type":        "string",
							"description": "The file path to be create, must be **absolute** path including file name and file extension",
						},
						"content": {
							"type":        "string",
							"description": "The **complete** text or source code string to be written into the file. Do not wrap this in markdown code blocks.",
						},
					},
				},
			},
		},
	}
	return t
}

// Run runs the tool with arguments
func (t WriteFile) Run(args any) (any, error) {
	in, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("wrong input")
	}

	path := in["path"].(string)
	content := in["content"].(string)

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Sprintf("failed to write file with error: %s", err.Error()), nil
	}

	return fmt.Sprintf("successfully write file to %s", path), nil
}
