package patchfile

import (
	"fmt"
	"os/exec"

	"github.com/nhv96/goas/pkg/tools"
)

// patchFile is a tool to patch content of a file without rewriting the whole file
type patchFile struct {
	tools.Base
}

// Init returns a tooler
func Init() tools.Tooler {
	t := patchFile{
		Base: tools.Base{
			Type: "function",
			Function: tools.Function{
				Name:        "patch_file",
				Description: "A tool to search and replace content of a file.",
				Parameters: tools.Parameters{
					Type:     "object",
					Required: []string{"path", "old", "new", "first"},
					Properties: map[string]map[string]any{
						"path": {
							"type":        "string",
							"description": "The file path, must be **absolute** path including file name and file extension.",
						},
						"old": {
							"type":        "string",
							"description": "The content to search for in the file to patch.",
						},
						"new": {
							"type":        "string",
							"description": "The new content to patch replacing the old content.",
						},
						"first": {
							"type":        "boolean",
							"description": "Replace only the first occurrence if 'true', default 'false'.",
						},
					},
				},
			},
		},
	}
	return t
}

// Run runs the tool with arguments
func (t patchFile) Run(args any) (any, error) {
	in, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("wrong input")
	}

	path := in["path"].(string)
	old := in["old"].(string)
	new := in["new"].(string)
	first := in["first"].(bool)

	var cmdStr string
	cmdStr = fmt.Sprintf("s/%s/%s/", old, new)

	if !first {
		cmdStr = fmt.Sprintf("s/%s/%s/g", old, new)
	}

	cmd := exec.Command("sed", "-i", cmdStr, path)

	fmt.Println("Running command", cmd.String())
	_, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("failed to patch file %s: %s", path, err.Error()), nil
	}

	return fmt.Sprintf("successfully patched file %s", path), nil
}
