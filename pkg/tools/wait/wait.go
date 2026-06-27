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
				Description: "A tool used to pause the program for a duration.",
				Parameters: tools.Parameters{
					Type:     "object",
					Required: []string{"t"},
					Properties: map[string]map[string]any{
						"t": {
							"type":        "integer",
							"description": "The time in milliseconds to wait for.",
						},
					},
				},
			},
		},
	}
	return t
}

// Run runs the tool with arguments, waiting for the specified duration.
func (t Wait) Run(args any) (any, error) {
	in, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("wait tool received invalid input type: expected map[string]any, got %T", args)
	}

	timeMsRaw, ok := in["t"].(float64)
	if !ok {
		return fmt.Sprintf("wait tool received invalid parameter 't': expected float64, got %T", in["t"]), nil
	}

	// Convert milliseconds to time.Duration
	duration := time.Duration(timeMsRaw * float64(time.Millisecond))

	time.Sleep(duration)

	return fmt.Sprintf("Finished waiting for %v", duration), nil
}
