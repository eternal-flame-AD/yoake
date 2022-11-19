package funcmap

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

func FindWord(s string, target ...string) bool {
	s = strings.ToLower(s)
	fields := strings.Fields(s)
	nonChar := regexp.MustCompile("[^a-z]")
	for i := range fields {
		fields[i] = nonChar.ReplaceAllString(fields[i], "")
	}
	s = " " + strings.Join(fields, " ") + " "
	for _, t := range target {
		if strings.Contains(s, " "+t+" ") {
			return true
		}
	}
	return false
}

func Contain(slice, target reflect.Value) (bool, error) {
	if slice.Kind() != reflect.Slice {
		return false, fmt.Errorf("Contain: slice is not a slice")
	}

	if slice.Type().Elem() != target.Type() {
		return false, fmt.Errorf("Contain: slice and target are not the same type")
	}

	for i := 0; i < slice.Len(); i++ {
		if slice.Index(i).Interface() == target.Interface() {
			return true, nil
		}
	}
	return false, nil
}
