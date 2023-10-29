package util

import (
	"fmt"
	"strings"
)

func IntArrayToString(a []int, delimiter string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delimiter, -1), "[]")
}
