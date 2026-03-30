package main

import "strings"

func ConvertPath(hertzPath string) string {
	segments := strings.Split(hertzPath, "/")
	for i, seg := range segments {
		if strings.HasPrefix(seg, ":") {
			segments[i] = "{" + seg[1:] + "}"
		} else if strings.HasPrefix(seg, "*") {
			segments[i] = "{" + seg[1:] + "}"
		}
	}
	return strings.Join(segments, "/")
}
