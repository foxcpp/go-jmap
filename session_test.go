package jmap

import (
	"encoding/json"
	"testing"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

var sessionBlob = `{
  "capabilities": {
    "urn:ietf:params:jmap:core": {
      "maxSizeUpload": 50000000,
      "maxConcurrentUpload": 8,
      "maxSizeRequest": 10000000,
      "maxConcurrentRequest": 8,
      "maxCallsInRequest": 32,
      "maxObjectsInGet": 256,
      "maxObjectsInSet": 128,
      "collationAlgorithms": [
        "i;ascii-numeric",
        "i;ascii-casemap",
        "i;unicode-casemap"
      ]
    },
    "urn:ietf:params:jmap:mail": {},
    "urn:ietf:params:jmap:contacts": {},
    "https://example.com/apis/foobar": {
      "maxFoosFinangled": 42
    }
  },
  "accounts": {
    "A13824": {
      "name": "john@example.com",
      "isPersonal": true,
      "isReadOnly": false,
      "accountCapabilities": {
        "urn:ietf:params:jmap:mail": {
          "maxMailboxesPerEmail": null,
          "maxMailboxDepth": 10
        },
        "urn:ietf:params:jmap:contacts": {
        }
      }
    },
    "A97813": {
      "name": "jane@example.com",
      "isPersonal": false,
      "isReadOnly": true,
      "accountCapabilities": {
        "urn:ietf:params:jmap:mail": {
          "maxMailboxesPerEmail": 1,
          "maxMailboxDepth": 10
        }
      }
    }
  },
  "primaryAccounts": {
    "urn:ietf:params:jmap:mail": "A13824",
    "urn:ietf:params:jmap:contacts": "A13824"
  },
  "username": "john@example.com",
  "apiUrl": "https://jmap.example.com/api/",
  "downloadUrl": "https://jmap.example.com/download/{accountId}/{blobId}/{name}?accept={type}",
  "uploadUrl": "https://jmap.example.com/upload/{accountId}/",
  "eventSourceUrl": "https://jmap.example.com/eventsource/?types={types}&closeafter={closeafter}&ping={ping}",
  "state": "75128aab4b1b"
}`

func TestSessionUnmarshal(t *testing.T) {
	s := Session{}
	assert.NilError(t, json.Unmarshal([]byte(sessionBlob), &s), "json.Unmarshal")

	assert.Check(t, cmp.Equal(UnsignedInt(50000000), s.CoreCapability.MaxSizeUpload))
	assert.Check(t, cmp.Equal("john@example.com", s.Accounts["A13824"].Name))
}

func TestSessionMarshal(t *testing.T) {
	s := Session{}
	assert.NilError(t, json.Unmarshal([]byte(sessionBlob), &s), "json.Unmarshal")

	blob, err := json.MarshalIndent(s, "", "  ")
	assert.NilError(t, err, "json.Marshal")

	// We can't just compare []byte because order of fields may be different.
	var original, remarshaled map[string]interface{}
	assert.NilError(t, json.Unmarshal([]byte(sessionBlob), &original))
	assert.NilError(t, json.Unmarshal(blob, &remarshaled))

	assert.Check(t, cmp.DeepEqual(original, remarshaled))
}
