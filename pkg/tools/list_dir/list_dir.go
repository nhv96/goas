package listdir

import (
	"fmt"
	"os/exec"

	"github.com/nhv96/goas/pkg/tools"
)

// listDir is a tool to list all files and folders in a specified directory path
type listDir struct {
	tools.Base
}

// Init returns a tooler
func Init() tools.Tooler {
	t := listDir{
		Base: tools.Base{
			Type: "function",
			Function: tools.Function{
				Name:        "list_dir",
				Description: "A tool to search for a file name or list all files and folders in a specified directory.",
				Parameters: tools.Parameters{
					Type:     "object",
					Required: []string{"path", "file_name"},
					Properties: map[string]map[string]any{
						"dir": {
							"type":        "string",
							"description": "The directory path to list files and folders, must be **absolute** path.",
						},
						"file_name": {
							"type":        "string",
							"description": "The keyword of file name or exact file name to search for.",
						},
					},
				},
			},
		},
	}
	return t
}

// Run runs the tool with arguments
func (t listDir) Run(args any) (any, error) {
	in, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("wrong input")
	}

	dir := in["dir"].(string)

	fileName := ""
	if _, ok := in["file_name"].(string); ok {
		fileName = in["file_name"].(string)
	}

	exculeDir := dir + "/.git"

	cmd := exec.Command("find", dir,
		"-path", exculeDir, "-prune", "-o",
		"-iname", fmt.Sprintf("*%s*", fileName),
		"-type", "f", "-print")

	fmt.Println("Running command", cmd.String())

	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("failed to list files in \"%s\": %s", dir, err.Error()), nil
	}
	fmt.Println(string(output))

	return string(output), nil
}
