package jmap

import (
	"encoding/json"
	"errors"
)

const CoreCapabilityName = "urn:ietf:params:jmap:core"

type CollationAlgo string

const (
	// The ASCIINumeric collation is a simple collation intended for use
	// with arbitrary sized unsigned decimal integer numbers stored as octet
	// strings. US-ASCII digits (0x30 to 0x39) represent digits of the numbers.
	// Before converting from string to integer, the input string is truncated
	// at the first non-digit character. All input is valid; strings which do
	// not start with a digit represent positive infinity.
	//
	// Defined in RFC 4790.
	ASCIINumeric CollationAlgo = "i;ascii-numeric"

	// The ASCIICasemap collation is a simple collation which operates on
	// octet strings and treats US-ASCII letters case-insensitively. It provides
	// equality, substring and ordering operations. All input is valid. Note that
	// letters outside ASCII are not treated case- insensitively.
	//
	// Defined in RFC 4790.
	ASCIICasemap = "i;ascii-casemap"

	// The "i;unicode-casemap" collation is a simple collation which is
	// case-insensitive in its treatment of characters. It provides equality,
	// substring, and ordering operations. The validity test operation returns "valid"
	// for any input.
	//
	// This collation allows strings in arbitrary (and mixed) character sets,
	// as long as the character set for each string is identified and it is
	// possible to convert the string to Unicode. Strings which have an
	// unidentified character set and/or cannot be converted to Unicode are not
	// rejected, but are treated as binary.
	//
	// Defined in RFC 5051.
	UnicodeCasemap = "i;unicode-casemap"

	// Octet collation is left out intentionally: "Protocols that want to make
	// this collation available have to do so by explicitly allowing it. If not
	// explicitly allowed, it MUST NOT be used."
)

type CoreCapability struct {
	// The maximum file size, in octets, that the server will accept for a
	// single file upload (for any purpose).
	MaxSizeUpload UnsignedInt `json:"maxSizeUpload"`

	// The maximum number of concurrent requests the server will accept to the
	// upload endpoint.
	MaxConcurrentUpload UnsignedInt `json:"maxConcurrentUpload"`

	// The maximum size, in octets, that the server will accept for a single
	// request to the API endpoint.
	MaxSizeRequest UnsignedInt `json:"maxSizeRequest"`

	// The maximum number of concurrent requests the server will accept to the
	// API endpoint.
	MaxConcurrentRequests UnsignedInt `json:"maxConcurrentRequests"`

	// The maximum number of method calls the server will accept in a single
	// request to the API endpoint.
	MaxCallsInRequest UnsignedInt `json:"maxCallsInRequest"`

	// The maximum number of objects that the client may request in a single
	// /get type method call.
	MaxObjectsInGet UnsignedInt `json:"maxObjectsInGet"`

	// The maximum number of objects the client may send to create, update or
	// destroy in a single /set type method call. This is the combined total, e.g.
	// if the maximum is 10 you could not create 7 objects and destroy 6, as this
	// would be 13 actions, which exceeds the limit.
	MaxObjectsInSet UnsignedInt `json:"maxObjectsInSet"`

	// A list of identifiers for algorithms registered in the collation
	// registry defined in RFC 4790 that the server supports for sorting
	// when querying records.
	CollationAlgorithms []CollationAlgo `json:"collationAlgorithms"`
}

// An account is a collection of data. A single account may contain an
// arbitrary set of data types, for example a collection of mail, contacts and
// calendars.
//
// See draft-ietf-jmap-core-17, section 1.6.2 for details.
// The documentation is taked from draft-ietf-jmap-core-17, section 2.
type Account struct {
	// A user-friendly string to show when presenting content from this
	// account, e.g. the email address representing the owner of the account.
	Name string `json:"name"`

	// This is true if the account belongs to the authenticated user, rather
	// than a group account or a personal account of another user that has been
	// shared with them.
	IsPersonal bool `json:"isPersonal"`

	// This is true if the entire account is read-only.
	IsReadOnly bool `json:"isReadOnly"`

	// The set of capability URIs for the methods supported in this account.
	// Each key is a URI for a capability that has methods you can use with
	// this account. The value for each of these keys is an object with further
	// information about the account’s permissions and restrictions with
	// respect to this capability, as defined in the capability’s
	// specification.
	Capabilities map[string]json.RawMessage `json:"accountCapabilities"`
}

// The Session object ... FIXME
//
// The documentation is taked from draft-ietf-jmap-core-17, section 2.
type Session struct {
	// An object specifying the capabilities of this server. Each key is a URI
	// for a capability supported by the server. The value for each of these
	// keys is an object with further information about the server’s
	// capabilities in relation to that capability.
	Capabilities map[string]json.RawMessage `json:"capabilities"`

	// Deserialized urn:ietf:params:jmap:core capability object.
	CoreCapability CoreCapability `json:"-"`

	// A map of account id to Account object for each account the user has
	// access to.
	Accounts map[ID]Account `json:"accounts"`

	// A map of capability URIs (as found in Capabilities) to the
	// account id to be considered the user’s main or default account for data
	// pertaining to that capability.
	PrimaryAccounts map[string]ID `json:"primaryAccounts"`

	// The username associated with the given credentials, or the empty string
	// if none.
	Username string `json:"username"`

	// The URL to use for JMAP API requests.
	APIURL string `json:"apiUrl"`

	// The URL endpoint to use when downloading files, in RFC 6570 URI
	// Template (level 1) format.
	DownloadURL string `json:"downloadUrl"`

	// The URL endpoint to use when uploading files, in RFC 6570 URI
	// Template (level 1) format.
	UploadURL string `json:"uploadUrl"`

	// The URL to connect to for push events, as described in section 7.3, in
	// RFC 6570 URI Template (level 1) format.
	EventSourceURL string `json:"eventSourceUrl"`

	// A string representing the state of this object on the server. If the
	// value of any other property on the session object changes, this string
	// will change.
	//
	// The current value is also returned on the API Response object, allowing
	// clients to quickly determine if the session information has changed
	// (e.g. an account has been added or removed) and so they need to refetch
	// the object.
	State string `json:"state"`
}

var ErrNoCoreCapability = errors.New("jmap: urn:ietf:params:jmap:core capability object is missing")

type session Session

func (s *Session) UnmarshalJSON(data []byte) error {
	raw := (*session)(s)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	coreCap, ok := raw.Capabilities[CoreCapabilityName]
	if !ok {
		return ErrNoCoreCapability
	}

	if err := json.Unmarshal(coreCap, &s.CoreCapability); err != nil {
		return err
	}
	return nil
}
