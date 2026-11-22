package agent

import (
	"context"
	"encoding/json"
	"fmt"
)

type ToolFunc struct {
	name        string
	description string
	parameters  json.RawMessage
	fn          func(ctx context.Context, ec *ExecutionContext, args json.RawMessage) (ToolResult, error)
}

func NewTool(name, description string, parameters json.RawMessage, fn func(ctx context.Context, ec *ExecutionContext, args json.RawMessage) (ToolResult, error)) Tool {
	return &ToolFunc{name: name, description: description, parameters: parameters, fn: fn}
}

func (t *ToolFunc) Name() string                { return t.name }
func (t *ToolFunc) Description() string         { return t.description }
func (t *ToolFunc) Parameters() json.RawMessage { return t.parameters }

func (t *ToolFunc) Execute(ctx context.Context, ec *ExecutionContext, args json.RawMessage) (ToolResult, error) {
	return t.fn(ctx, ec, args)
}

// Registry is an explicit allowlist: only tools registered here can be called
// by the model. Duplicate names are rejected at construction.
type Registry struct {
	byName map[string]Tool
	order  []string
}

func NewRegistry(tools ...Tool) (*Registry, error) {
	r := &Registry{byName: make(map[string]Tool, len(tools))}
	for _, tool := range tools {
		if tool == nil {
			continue
		}
		name := tool.Name()
		if name == "" {
			return nil, fmt.Errorf("agent: tool name is empty")
		}
		if _, exists := r.byName[name]; exists {
			return nil, fmt.Errorf("agent: duplicate tool %q", name)
		}
		r.byName[name] = tool
		r.order = append(r.order, name)
	}
	return r, nil
}

func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.byName[name]
	return tool, ok
}

func (r *Registry) Specs() []ToolSpec {
	specs := make([]ToolSpec, 0, len(r.order))
	for _, name := range r.order {
		tool := r.byName[name]
		params := tool.Parameters()
		if len(params) == 0 {
			params = json.RawMessage(`{"type":"object","properties":{}}`)
		}
		specs = append(specs, ToolSpec{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  params,
		})
	}
	return specs
}

func (r *Registry) Names() []string {
	out := make([]string, len(r.order))
	copy(out, r.order)
	return out
}
