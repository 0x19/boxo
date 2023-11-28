package attribute

import "fmt"

func StringSlice(k string, v []string) struct{} {
	return struct{}{}
}

func String(k string, v string) struct{} {
	return struct{}{}
}

func Stringer(k string, v fmt.Stringer) struct{} {
	return struct{}{}
}

func Int(k string, v int) struct{} {
	return struct{}{}
}
