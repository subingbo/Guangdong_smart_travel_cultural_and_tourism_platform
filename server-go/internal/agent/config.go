package agent

import (
	"os"
	"strconv"
	"time"
)

// Limits bounds one run. Defaults are intentionally small: the production
// host is a 2C1G VM with a 256-384MB cgroup.
type Limits struct {
	MaxSteps            int
	MaxToolCallsPerStep int
	ToolTimeout         time.Duration
	MaxArgumentsBytes   int
	MaxToolResultBytes  int
	ToolBudgets         map[string]int
}

func DefaultLimits() Limits {
	return Limits{
		MaxSteps:            8,
		MaxToolCallsPerStep: 3,
		ToolTimeout:         20 * time.Second,
		MaxArgumentsBytes:   16 * 1024,
		MaxToolResultBytes:  16 * 1024,
		ToolBudgets: map[string]int{
			"search_places":    4,
			"get_place":        8,
			"list_favorites":   2,
			"search_knowledge": 4,
			"get_weather":      2,
			"route_leg":        24,
			"plan_itinerary":   3,
			"submit_plan":      2,
			"list_my_routes":   2,
		},
	}
}

// LoadLimits reads AGENT_* overrides and keeps the defaults for absent keys.
func LoadLimits() Limits {
	limits := DefaultLimits()
	limits.MaxSteps = envInt("AGENT_MAX_STEPS", limits.MaxSteps)
	limits.MaxToolCallsPerStep = envInt("AGENT_MAX_TOOL_CALLS_PER_STEP", limits.MaxToolCallsPerStep)
	limits.ToolTimeout = time.Duration(envInt("AGENT_TOOL_TIMEOUT_SECONDS", int(limits.ToolTimeout/time.Second))) * time.Second
	limits.MaxArgumentsBytes = envInt("AGENT_MAX_ARGUMENTS_BYTES", limits.MaxArgumentsBytes)
	limits.MaxToolResultBytes = envInt("AGENT_MAX_TOOL_RESULT_BYTES", limits.MaxToolResultBytes)
	return limits
}

func (l Limits) Budget(name string) int {
	if n, ok := l.ToolBudgets[name]; ok {
		return n
	}
	return 99
}

func (l Limits) Clone() Limits {
	out := l
	out.ToolBudgets = make(map[string]int, len(l.ToolBudgets))
	for k, v := range l.ToolBudgets {
		out.ToolBudgets[k] = v
	}
	return out
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
