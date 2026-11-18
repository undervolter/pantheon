package domain

import "sort"

func sortedCopy(list []string) []string {
	out := make([]string, 0, len(list))
	for _, v := range list {
		if v != "" {
			out = append(out, v)
		}
	}
	sort.Strings(out)
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
