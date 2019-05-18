package jmap

import (
	"encoding/json"
	"testing"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

const requestErr = `{
  "type": "urn:ietf:params:jmap:error:unknownCapability",
  "status": 400,
  "detail": "The request object used capability 'https://example.com/apis/foobar', which is not supported by this server."
}`

const requestLimitErr = `{
  "type": "urn:ietf:params:jmap:error:limit",
  "limit": "maxSizeRequest",
  "status": 400,
  "detail": "The request is larger than the server is willing to process."
}`

func TestRequestErrorUnmarshal(t *testing.T) {
	errObj := RequestError{}
	err := json.Unmarshal([]byte(requestErr), &errObj)
	assert.NilError(t, err, "json.Unmarshal")

	assert.Check(t, cmp.Equal(400, errObj.Status))
	assert.Check(t, cmp.Equal(ProblemUnknownCapability, errObj.Type))

	t.Run("properties", func(t *testing.T) {
		errObj := RequestError{}
		err := json.Unmarshal([]byte(requestLimitErr), &errObj)
		assert.NilError(t, err, "json.Unmarshal")

		assert.Check(t, cmp.Equal(400, errObj.Status))
		assert.Check(t, cmp.Equal(ProblemLimit, errObj.Type))
		assert.Check(t, cmp.Equal(errObj.Properties["limit"], "maxSizeRequest"))
	})
}

func TestRequestErrorMarshal(t *testing.T) {
	errObj := RequestError{}
	errObj.Type = ProblemUnknownCapability
	errObj.Status = 400
	errObj.Detail = "something is broken, yay!"

	blob, err := json.Marshal(errObj)
	assert.NilError(t, err, "json.Marshal")
	assert.Check(t, cmp.Equal(`{"detail":"something is broken, yay!","status":400,"type":"urn:ietf:params:jmap:error:unknownCapability"}`, string(blob)))

	t.Run("properties", func(t *testing.T) {
		errObj := RequestError{}
		errObj.Type = ProblemLimit
		errObj.Status = 400
		errObj.Detail = "something is broken, yay!"
		errObj.Properties = map[string]interface{}{
			"limit": "maxSizeRequest",
		}

		blob, err := json.Marshal(errObj)
		assert.NilError(t, err, "json.Marshal")
		assert.Check(t, cmp.Equal(`{"detail":"something is broken, yay!","limit":"maxSizeRequest","status":400,"type":"urn:ietf:params:jmap:error:limit"}`, string(blob)))
	})
}
