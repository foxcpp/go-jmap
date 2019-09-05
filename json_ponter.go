package jmap

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var jsonPtrRegexp = regexp.MustCompile(`(/(([^/~])|(~[01]))*)`)

// ErrInvalidPointerSyntax is returned when the json pointer syntax does not
// conform to https://tools.ietf.org/html/rfc6901#page-2
var ErrInvalidPointerSyntax = errors.New("invalid json pointer syntax")

// ErrNoPointerValue is returned when there is no value found on the json ponter
// path.
var ErrNoPointerValue = errors.New("no pointer value")

// JSONObject is an interface for retrieving object properties based on string
// keys
type JSONObject interface {
	Get(key string) JSONObject
	Interface() interface{}
}

type jsonPtrWrapper struct {
	value reflect.Value
}

func (j *jsonPtrWrapper) Interface() interface{} {
	return j.value.Interface()
}

func (j *jsonPtrWrapper) Get(key string) JSONObject {
	v := j.value
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Map:
		r := v.MapRange()
		for r.Next() {
			k := r.Key()
			v := r.Value()
			if k.String() == key {
				return &jsonPtrWrapper{value: v}
			}
		}
	case reflect.Slice:
		i, err := strconv.Atoi(key)
		if err != nil {
			return nil
		}
		if i >= v.Len() {
			// out of range
			return nil
		}
		return &jsonPtrWrapper{value: v.Index(i)}
	case reflect.Struct:
		typ := v.Type()
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			ch, _ := utf8.DecodeRuneInString(f.Name)
			if !unicode.IsUpper(ch) {
				continue
			}
			if f.Name == key {
				return &jsonPtrWrapper{value: v.Field(i)}
			}
			tag := f.Tag.Get("json")
			if tag != "" {
				tag = strings.Split(tag, ",")[0]
				if tag == key {
					return &jsonPtrWrapper{value: v.Field(i)}
				}
			}
		}
	}
	return nil
}

// GetJSON uses rfc6901 to return object from path which is a json pointer
// string syntax.
//
// object can either be a struct/array/map. It is possible to pass anything that
// implements JSONObject interface as object.
func GetJSON(path string, object interface{}) (interface{}, error) {
	if !jsonPtrRegexp.Match([]byte(path)) {
		return nil, ErrInvalidPointerSyntax
	}
	var o JSONObject
	if x, ok := object.(JSONObject); ok {
		o = x
	} else {
		o = &jsonPtrWrapper{value: reflect.ValueOf(object)}
	}

	parts := jsonPtrRegexp.FindAllString(path, -1)
	for _, p := range parts {
		if o == nil {
			return nil, ErrNoPointerValue
		}
		p = strings.Replace(p, "~1", "/", -1)
		p = strings.Replace(p, "~0", "~", -1)
		if p[0] == '/' {
			p = p[1:]
		}
		o = o.Get(p)
	}
	if o == nil {
		return nil, ErrNoPointerValue
	}
	return o.Interface(), nil
}
