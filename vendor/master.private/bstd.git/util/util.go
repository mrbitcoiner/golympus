package util

import "strings"

func Dirname(arg string) string {
	slashIdx := strings.LastIndex(arg, "/")
	if slashIdx < 0 {
		return "."
	}
	if slashIdx == 0 {
		return "/"
	}
	return arg[:slashIdx]
}
