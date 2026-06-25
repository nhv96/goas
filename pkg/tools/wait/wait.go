package wait

import (
	"fmt"
	"time"

	"github.com/nhv96/goas/pkg/tools"
)

// Wait is a tool that make the program pause/wait for a duration
type Wait struct {
	tools.Base
}

// Init returns a tooler
func Init() tools.Tooler {
	t := Wait{
		Base: tools.Base{
			Type: "function",
			Function: tools.Function{
				Name:        "wait",
				Description: "Wait for a duration then continue the program.",
				Parameters: tools.Parameters{
					Type:     "object",
					Required: []string{"t"},
					Properties: map[string]map[string]any{
						"t": {
							"type":        "integer",
							"description": "the time in milliseconds to wait",
						},
					},
				},
			},
		},
	}
	return t
}

// Run runs the tool with arguments
func (t Wait) Run(args any) (any, error) {
	in, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("wrong input")
	}

	timeToWait := in["t"].(float64)

	time.Sleep(time.Duration(timeToWait) * time.Millisecond)

	return fmt.Sprintf("Finished waiting for %f milliseconds", timeToWait), nil
}
