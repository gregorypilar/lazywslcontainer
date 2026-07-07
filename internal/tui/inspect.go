package tui

import (
	"encoding/json"
	"fmt"
	"strings"
)

func prettyInspect(raw string) string {
	var data any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return raw
	}
	if arr, ok := data.([]any); ok && len(arr) == 1 {
		data = arr[0]
	}
	var sb strings.Builder
	renderJSON(&sb, data, 0, false)
	return sb.String()
}

func renderJSON(sb *strings.Builder, v any, indent int, inArray bool) {
	pad := strings.Repeat("  ", indent)
	switch t := v.(type) {
	case map[string]any:
		if inArray {
			sb.WriteString("\n")
		}
		keys := sortedKeys(t)
		sb.WriteString("{\n")
		for i, k := range keys {
			sb.WriteString(pad + "  ")
			sb.WriteString(fmt.Sprintf("%q: ", k))
			renderJSON(sb, t[k], indent+1, false)
			if i < len(keys)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString(pad + "}")
	case []any:
		if len(t) == 0 {
			sb.WriteString("[]")
			return
		}
		sb.WriteString("[")
		for i, item := range t {
			renderJSON(sb, item, indent+1, true)
			if i < len(t)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString("]")
	case string:
		sb.WriteString(fmt.Sprintf("%q", t))
	case float64:
		sb.WriteString(fmt.Sprintf("%g", t))
	case bool:
		sb.WriteString(fmt.Sprintf("%t", t))
	case nil:
		sb.WriteString("null")
	default:
		sb.WriteString(fmt.Sprintf("%v", t))
	}
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
