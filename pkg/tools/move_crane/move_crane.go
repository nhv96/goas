package movecrane

import (
	"fmt"

	"github.com/nhv96/goas/pkg/tools"
)

// MoveCrane is a tool to send signal to modbus Server and move the crane
// in simulation
type MoveCrane struct {
	tools.Base
}

// Init returns a tooler
func Init() tools.Tooler {
	t := MoveCrane{
		Base: tools.Base{
			Type: "function",
			Function: tools.Function{
				Name:        "move_crane",
				Description: "Send order to move the crane to x and y position",
				Parameters: tools.Parameters{
					Type:     "object",
					Required: []string{"x", "y"},
					Properties: map[string]map[string]any{
						"x": {
							"type":        "integer",
							"description": "the input of x coordinate",
						},
						"y": {
							"type":        "integer",
							"description": "the input of y coordinate",
						},
					},
				},
			},
		},
	}
	return t
}

// Run runs the tool with arguments
// func (t MoveCrane) Run(args any) (any, error) {
// 	in, ok := args.(map[string]any)
// 	if !ok {
// 		return nil, fmt.Errorf("wrong input")
// 	}

// 	x := in["x"].(float64)
// 	y := in["y"].(float64)

// 	w := mb.Init()
// 	defer w.Close()

// 	xWriteAddr := 1
// 	yWriteAddr := 2

// 	yReadAddr := 16

// 	// move by y-direction first
// 	err := w.WriteAddress(yWriteAddr, int(y))
// 	if err != nil {
// 		return fmt.Sprintf("error write y: %s", err.Error()), nil
// 	}

// 	timeout := 10 * time.Second
// 	timeoutChan := time.After(timeout)

// 	// keep polling until crane finished moving by y-direction
// 	ticker := time.NewTicker(50 * time.Millisecond)
// 	defer ticker.Stop()

// Loop:
// 	for {
// 		select {
// 		case <-timeoutChan:
// 			return "error timeout while polling", nil
// 		case <-ticker.C:
// 			currpos, err := w.ReadAddress(yReadAddr)
// 			if err != nil {
// 				return fmt.Sprintf("error read y: %s", err.Error()), nil
// 			}

// 			if currpos == int(y) {
// 				break Loop
// 			}
// 		}
// 	}

// 	// then move by x-direction
// 	err = w.WriteAddress(xWriteAddr, int(x))
// 	if err != nil {
// 		return fmt.Sprintf("error write x: %s", err.Error()), nil
// 	}

// 	// maybe no need polling here

// 	return fmt.Sprintf("Moved crane to %f, %f", x, y), nil
// }

func (t MoveCrane) Run(args any) (any, error) {
	in, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("wrong input")
	}

	x := in["x"].(float64)
	y := in["y"].(float64)

	return fmt.Sprintf("Moved crane to %f, %f", x, y), nil
}
