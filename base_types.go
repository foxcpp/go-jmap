package jmap

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"time"
	"unicode"
)

/*
This file defines fundamental data types used in the JMAP specification.

Since some of the types have specific constraints, Go native types are not used
directly but instead wrapped into containers that enforce these constraints.

Type constraints checks are enforced during conversion to/from JSON.
*/

var ErrOutOfRange = errors.New("jmap: integer value is not within allowed range")

// Int type is an integer in the range -2^53+1 <= value <= 2^53-1.
type Int int64

// Valid checks whether value Int is set to is within the allowed range.
func (i Int) Valid() bool {
	return i >= (-2<<52+1) && i <= (2<<52-1)
}

func (i Int) MarshalJSON() ([]byte, error) {
	if !i.Valid() {
		return nil, ErrOutOfRange
	}

	return json.Marshal(int64(i))
}

func (i *Int) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*int64)(i)); err != nil {
		return err
	}
	if !i.Valid() {
		return ErrOutOfRange
	}
	return nil
}

// UnsignedInt is an integer in the range 0 <= value <= 2^53-1.
type UnsignedInt uint64

// Valid checks whether value UnsignedInt is set to is within the allowed
// range.
func (i UnsignedInt) Valid() bool {
	return i <= (2<<52 - 1)
}

func (i UnsignedInt) MarshalJSON() ([]byte, error) {
	if !i.Valid() {
		return nil, ErrOutOfRange
	}

	return json.Marshal(uint64(i))
}

func (i *UnsignedInt) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*uint64)(i)); err != nil {
		return err
	}
	if !i.Valid() {
		return ErrOutOfRange
	}
	return nil
}

// Date is a time.Time that is serialized to JSON in RFC 3339 format (without
// the fractional part).
type Date time.Time

func (d Date) MarshalText() ([]byte, error) {
	b := make([]byte, 0, len(time.RFC3339))
	b = time.Time(d).AppendFormat(b, time.RFC3339)
	return b, nil
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	var err error
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*(*time.Time)(d), err = time.Parse(time.RFC3339, s)
	return err
}

// Date is a time.Time that is serialized to JSON in RFC 3339 format (without
// the fractional part) in UTC timezone.
//
// If UTCDate value is not in UTC, it will be converted to UTC during
// serialization.
type UTCDate time.Time

func (d UTCDate) MarshalText() ([]byte, error) {
	b := make([]byte, 0, len(time.RFC3339)+2)
	b = time.Time(d).UTC().AppendFormat(b, time.RFC3339)
	return b, nil
}

func (d *UTCDate) UnmarshalJSON(data []byte) error {
	var s string
	var err error
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*(*time.Time)(d), err = time.ParseInLocation(time.RFC3339, s, time.UTC)
	return err
}

// Id is a string of at least 1 and maximum 255 octets in size that must
// contain only characters from the "URL and Filename safe" Base 64 Alphabet
// (see section 5 of RFC 4648), excluding the pad character.
// This means the allowed charcters are ASCII alphanumeric characters, hypen
// and underscore.
type ID string

var ErrInvalidId = errors.New("jmap: invalid id")

var validIdRegexp = regexp.MustCompile(`^[A-Za-z0-9\-_]+$`)

// Safe function checks whether Id value is safe to use in filesystems, URLs,
// etc without escaping. In particular, for Id to be considered safe it:
// - Should not start with a dash
// - Should not start with digits
// - Should not contain only digits.
// - Should not be equal to "NIL".
//
// JMAP specification requires that all used Ids are safe ("SHOULD").
func (id ID) Safe() bool {
	if !id.Valid() {
		return false
	}

	if id[0] == '-' || unicode.IsDigit(rune(id[0])) {
		return false
	}

	onlyDigits := true
	for _, b := range id {
		if !unicode.IsDigit(b) {
			onlyDigits = false
		}
	}
	if onlyDigits {
		return false
	}

	if id == "NIL" {
		return false
	}

	return true
}

// Valid checks whether Id value is a valid identifier.
func (id ID) Valid() bool {
	if len(id) < 1 || len(id) > 255 {
		return false
	}

	return validIdRegexp.MatchString(string(id))
}

func (id ID) MarshalText() ([]byte, error) {
	if !id.Valid() {
		return nil, ErrInvalidId
	}

	return []byte(id), nil
}

func (id *ID) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*string)(id)); err != nil {
		return err
	}

	if !id.Valid() {
		return ErrInvalidId
	}
	return nil
}

// RandomID generates random valid & safe Id object.
func RandomID() (ID, error) {
	b := make([]byte, 32)
	var id ID
	for !(id.Valid() && id.Safe()) {
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		id = ID(hex.EncodeToString(b))
	}
	return id, nil
}
