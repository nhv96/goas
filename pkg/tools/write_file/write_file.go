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
				Description: "A tool to write content to a file.",
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
		return nil, fmt.Errorf("invalid input type: expected map[string]any, got %T", args)
	}

	pathVal, ok := in["path"].(string)
	if !ok || pathVal == "" {
		return "failed to write file, missing or invalid required parameter: 'path' must be a non-empty string", nil
	}

	contentVal, ok := in["content"].(string)
	if !ok || contentVal == "" {
		return "failed to write file, missing or invalid required parameter: 'content' must be a non-empty string", nil
	}

	// Execute the file write operation
	err := os.WriteFile(pathVal, []byte(contentVal), 0644)
	if err != nil {
		// Return the actual OS error directly
		return fmt.Sprintf("failed to write file to %s: %s", pathVal, err.Error()), nil
	}

	// Return a success message and nil error
	return fmt.Sprintf("successfully write file to %s", pathVal), nil
}
