package audit

import (
	"fmt"
	"sort"
)

func Walk(root any, visit func(Node)) {
	walk(root, "", "", nil, "", visit)
}

func walk(value any, path string, key string, parent any, parentPath string, visit func(Node)) {
	visit(Node{
		Path:       path,
		Key:        key,
		Value:      value,
		Parent:     parent,
		ParentPath: parentPath,
	})

	switch current := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(current))
		for childKey := range current {
			keys = append(keys, childKey)
		}
		sort.Strings(keys)

		for _, childKey := range keys {
			childPath := childKey
			if path != "" {
				childPath = path + "." + childKey
			}
			walk(current[childKey], childPath, childKey, current, path, visit)
		}

	case []any:
		for i, child := range current {
			childPath := fmt.Sprintf("[%d]", i)
			if path != "" {
				childPath = fmt.Sprintf("%s[%d]", path, i)
			}
			walk(child, childPath, "", current, path, visit)
		}
	}
}
