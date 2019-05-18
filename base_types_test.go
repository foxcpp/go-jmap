package jmap

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

func TestIntIsValid(t *testing.T) {
	assert.Check(t, !(Int(2 << 54).Valid()))
	assert.Check(t, !(Int(-2 << 54).Valid()))
	assert.Check(t, (Int(2 << 50).Valid()))
	assert.Check(t, (Int(-2 << 50).Valid()))
}

func TestIntMarshalJSON(t *testing.T) {
	_, err := json.Marshal(Int(2 << 54))
	assert.Check(t, cmp.ErrorContains(err, ErrOutOfRange.Error()), "json.Marshal")

	b, err := json.Marshal(Int(2 << 50))
	assert.NilError(t, err, "json.Marshal")
	assert.Check(t, cmp.DeepEqual(b, []byte("2251799813685248")))
}

func TestIntUnmarshalJSON(t *testing.T) {
	val := []byte("2251799813685248")
	var i Int
	err := json.Unmarshal(val, &i)
	assert.NilError(t, err, "json.Unmarshal")
	assert.Check(t, cmp.Equal(i, Int(2251799813685248)))

	val = []byte("225179981368524800")
	err = json.Unmarshal(val, &i)
	assert.Check(t, cmp.ErrorContains(err, ErrOutOfRange.Error()), "json.Unmarshal")
}
func TestUnsignedIntIsValid(t *testing.T) {
	assert.Check(t, !(UnsignedInt(2 << 54).Valid()))
	assert.Check(t, (UnsignedInt(2 << 50).Valid()))
}

func TestUnsignedIntMarshalJSON(t *testing.T) {
	_, err := json.Marshal(UnsignedInt(2 << 54))
	assert.Check(t, cmp.ErrorContains(err, ErrOutOfRange.Error()), "json.Marshal")

	b, err := json.Marshal(UnsignedInt(2 << 50))
	assert.NilError(t, err, "json.Marshal")
	assert.Check(t, cmp.DeepEqual(b, []byte("2251799813685248")))
}

func TestUnsignedIntUnmarshalJSON(t *testing.T) {
	val := []byte("2251799813685248")
	var i UnsignedInt
	err := json.Unmarshal(val, &i)
	assert.NilError(t, err, "json.Unmarshal")
	assert.Check(t, cmp.Equal(i, UnsignedInt(2251799813685248)))

	val = []byte("225179981368524800")
	err = json.Unmarshal(val, &i)
	assert.Check(t, cmp.ErrorContains(err, ErrOutOfRange.Error()), "json.Unmarshal")
}

func TestUTCDateMarshal(t *testing.T) {
	// Non-UTC time should be converted to UTC.
	d := UTCDate(time.Now())
	b, err := json.Marshal(d)
	assert.NilError(t, err, "json.Marshal")
	assert.Check(t, bytes.HasSuffix(b, []byte(`Z"`)), b)
}

func TestIdIsValid(t *testing.T) {
	assert.Check(t, ID("iamValid0_-").Valid())
	assert.Check(t, !ID("").Valid())
	assert.Check(t, !ID(strings.Repeat("A", 525)).Valid())
	assert.Check(t, !ID("i'mnot0").Valid())
	assert.Check(t, !ID("im bad too0").Valid())
}

func TestIdIsSafe(t *testing.T) {
	assert.Check(t, ID("imsafe").Safe())
	assert.Check(t, !ID("i'mnot valid").Safe())
	assert.Check(t, !ID("-000").Safe())
	assert.Check(t, !ID("000").Safe())
	assert.Check(t, !ID("0aaa").Safe())
	assert.Check(t, !ID("NIL").Safe())
}
