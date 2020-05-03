package util

import "strings"

/*
可以用path.IsAbs(..)来替换
还有个path.Join(..,..)也很好用
 */
func IsAbsolutePath(path string) bool {
	if strings.Contains(path, ":") || strings.HasPrefix(path, "/") {
		return true
	}
	return false
}