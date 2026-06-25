package agent

import "github.com/nhv96/goas/pkg/tools"

type ToolBag map[string]tools.Tooler

func (bag ToolBag) Descriptions() []string {
	desc := []string{}
	for _, tool := range bag {
		desc = append(desc, tool.String())
	}

	return desc
}
