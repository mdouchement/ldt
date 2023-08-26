package primitive

import (
	"os"
	"strings"
)

func LongestCommonPrefix(s1, s2 string) string {
	if len(s1) > len(s2) {
		s1, s2 = s2, s1
	}

	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return s1[:i]
		}
	}

	return s1
}

func LongestCommonPathPrefix(p1, p2 string) string {
	const sep = string(os.PathSeparator)

	path1 := strings.Split(p1, sep)
	path2 := strings.Split(p2, sep)
	if len(path1) > len(path2) {
		path1, path2 = path2, path1
	}

	for i := 0; i < len(path1); i++ {
		if path1[i] != path2[i] {
			return strings.Join(path1[:i], sep)
		}
	}

	return p1
}
