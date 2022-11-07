package funcmap

import (
	"fmt"
	"reflect"
)

func FuncGet(target reflect.Value, name string) (interface{}, error) {
	switch target.Kind() {
	case reflect.Map:
		if target.IsNil() {
			target = reflect.MakeMap(reflect.MapOf(target.Type().Key(), reflect.TypeOf("")))
		}
		v := target.MapIndex(reflect.ValueOf(name))
		if !v.IsValid() {
			return nil, nil
		}
		return v.Interface(), nil
	case reflect.Struct:
	case reflect.Interface:
		return Lookup(name, target)
	}

	return nil, fmt.Errorf("cannot get %s from type %v", name, target.Type())
}

func FuncSet(target reflect.Value, name string, value interface{}) (interface{}, error) {
	switch target.Kind() {
	case reflect.Map:
		target.SetMapIndex(reflect.ValueOf(name), reflect.ValueOf(value))
		return "", nil
	case reflect.Struct:
	case reflect.Interface:
		field := target.FieldByName(name)
		if field.IsValid() {
			field.Set(reflect.ValueOf(value))
			return "", nil
		}
	}

	return "", fmt.Errorf("cannot set %s to type %v", name, target.Type())
}
