package grep

import (
	"fmt"
	"os/exec"

	"github.com/nhv96/goas/pkg/tools"
)

type Grep struct {
	tools.Base
}

func Init() tools.Tooler {
	t := Grep{
		Base: tools.Base{
			Type: "function",
			Function: tools.Function{
				Name:        "grep",
				Description: "A tool to search a pattern accross files in a directory.",
				Parameters: tools.Parameters{
					Type:     "object",
					Required: []string{"pattern", "directory"},
					Properties: map[string]map[string]any{
						"pattern": {
							"type":        "string",
							"description": "The text string or keyword to search for.",
						},
						"directory": {
							"type":        "string",
							"description": "The path to directory, must be **absolute** path.",
						},
					},
				},
			},
		},
	}

	return t
}

func (t Grep) Run(args any) (any, error) {
	in, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("wrong input")
	}

	str := in["pattern"].(string)
	dir := in["directory"].(string)

	cmd := exec.Command("grep", "-rin", str, dir)

	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("failed to search for \"%s\" in \"%s\": %s", str, dir, err.Error()), nil
	}

	return string(output), nil
}
