package project

import "strings"

// These are opaque application owners already stored by ZhiYuan Agent, not
// filesystem paths. Keep every other path/URI subject to the upstream validator.
func isZhiYuanMemoryNamespace(name string) bool {
	if name == "personal://zhiyuan-agent/user" {
		return true
	}
	id, ok := strings.CutPrefix(name, "companion://zhiyuan-agent/")
	if !ok || len(id) == 0 || len(id) > 128 {
		return false
	}
	for _, ch := range id {
		if !(ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || ch == '-' || ch == '_') {
			return false
		}
	}
	return true
}
