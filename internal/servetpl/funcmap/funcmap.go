package funcmap

import (
	"fmt"
	"log"
	"reflect"
)

var ErrEarlyTermination = fmt.Errorf("template was early terminated by calling {{ stop }}")

func Stop() (string, error) {
	return "", ErrEarlyTermination
}

func Lookup(name string, args ...reflect.Value) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("lookup expects at least 1 argument, got 0")
	}
	zero := reflect.ValueOf("")

	target := args[0]
	log.Printf("lookup %s %v", name, target)
	if len(args) > 1 {
		zero = args[1]
	}
	switch target.Kind() {
	case reflect.Map:
		if target.IsNil() || target.Type().Elem().Kind() != zero.Kind() {
			target = reflect.MakeMap(reflect.MapOf(target.Type().Key(), zero.Type()))
			target.SetMapIndex(reflect.ValueOf(name), zero)
		}

		return target.MapIndex(reflect.ValueOf(name)).Interface(), nil
	case reflect.Pointer:
		field := target.MethodByName(name)
		if field.IsValid() {
			return field.Interface(), nil
		}
	case reflect.Struct:
		field := target.FieldByName(name)
		if field.IsValid() {
			return field.Interface(), nil
		}
	case reflect.Interface:
		method := target.MethodByName(name)
		if method.IsValid() {
			return method.Interface(), nil
		}
	default:
		return nil, fmt.Errorf("cannot lookup %s from type %v", name, target.Type())
	}
	return nil, fmt.Errorf("no such method or field %s", name)
}

func Invoke(name string, target reflect.Value, args ...reflect.Value) (any, error) {
	if name != "" {
		t, err := Lookup(name, target)
		if err != nil {
			return nil, err
		}
		target = reflect.ValueOf(t)
	}
	for i, arg := range args {
		log.Printf("invoke %s arg[%d]=%v", name, i, arg)
	}
	ret := target.Call(args)
	if len(ret) == 0 {
		return nil, nil
	}

	if err, ok := ret[len(ret)-1].Interface().(error); ok && err != nil {
		return nil, err
	}
	for i, r := range ret {
		log.Printf("invoke %s ret[%d]=%v", name, i, r)
	}

	switch len(ret) {
	case 0:
		return nil, nil
	case 1:
		return ret[0].Interface(), nil
	default:
		var rets []any
		for _, r := range ret {
			rets = append(rets, r.Interface())
		}
		return rets, nil
	}

}

func Void(args ...reflect.Value) string {
	return ""
}

func GetFuncMap() map[string]interface{} {
	return map[string]interface{}{
		"lookup":          Lookup,
		"invoke":          Invoke,
		"void":            Void,
		"get":             FuncGet,
		"set":             FuncSet,
		"math":            Math,
		"xml":             EscapeXML,
		"twilio_validate": TwilioValidate,
		"stop":            Stop,
		"trima_img":       TrimaImg,
		"parse_json":      ParseJSON,
		"json":            MarshalJSON,
		"get_auth":        AuthGet,
		"auth_login":      AuthLogin,
	}
}
