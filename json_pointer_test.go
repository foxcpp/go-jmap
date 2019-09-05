package jmap

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestJSONPointer(t *testing.T) {
	document := `   {
		"foo": ["bar", "baz"],
		"": 0,
		"a/b": 1,
		"c%d": 2,
		"e^f": 3,
		"g|h": 4,
		"i\\j": 5,
		"k\"l": 6,
		" ": 7,
		"m~n": 8
	 }`

	t.Run("map", func(ts *testing.T) {
		sample := []struct {
			path   string
			expect interface{}
		}{
			{path: "/foo", expect: []interface{}{"bar", "baz"}},
			{path: "/foo/0", expect: "bar"},
			{path: "/", expect: float64(0)},
			{path: "/a~1b", expect: float64(1)},
			{path: "/c%d", expect: float64(2)},
			{path: "/e^f", expect: float64(3)},
			{path: "/g|h", expect: float64(4)},
			{path: "/g|h", expect: float64(4)},
			{path: "/i\\j", expect: float64(5)},
			{path: "/k\"l", expect: float64(6)},
			{path: "/ ", expect: float64(7)},
			{path: "/m~0n", expect: float64(8)},
		}
		var o map[string]interface{}
		err := json.Unmarshal([]byte(document), &o)
		if err != nil {
			ts.Fatal(err)
		}
		for _, v := range sample {
			r, err := GetJSON(v.path, o)
			if err != nil {
				ts.Error(v.path, err)
				continue
			}
			if !reflect.DeepEqual(r, v.expect) {
				t.Errorf("%s : expected %T got %T %#v %#v", v.path, v.expect, r, v.expect, r)
			}
		}
	})
	t.Run("struct", func(ts *testing.T) {
		sample := []struct {
			path   string
			expect interface{}
		}{
			{path: "/foo", expect: []interface{}{"bar", "baz"}},
			{path: "/foo/0", expect: "bar"},
			{path: "/a~1b", expect: float64(1)},
		}
		var o = struct {
			F1 []interface{} `json:"foo"`
			F2 float64       `json:"a/b"`
		}{}
		err := json.Unmarshal([]byte(document), &o)
		if err != nil {
			ts.Fatal(err)
		}
		for _, v := range sample {
			r, err := GetJSON(v.path, o)
			if err != nil {
				ts.Error(v.path, err)
				continue
			}
			if !reflect.DeepEqual(r, v.expect) {
				t.Errorf("%s : expected %T got %T %#v %#v", v.path, v.expect, r, v.expect, r)
			}
		}
	})
}
