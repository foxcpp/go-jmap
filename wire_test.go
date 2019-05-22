package jmap

import (
	"encoding/json"
	"testing"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

func TestUnmarshalRawInvocation(t *testing.T) {
	blob := `["meth", {"arg1":1}, "callid"]`
	rawInv := rawInvocation{}
	err := json.Unmarshal([]byte(blob), &rawInv)
	assert.NilError(t, err, "UnmarshalInvocation")
	assert.Check(t, cmp.Equal("meth", rawInv.Name))
	assert.Check(t, cmp.Equal("callid", rawInv.CallID))
	assert.Check(t, cmp.DeepEqual(`{"arg1":1}`, string(rawInv.Args)))

	t.Run("missing call ID", func(t *testing.T) {
		blob := `["meth", {"arg1":1}]`
		rawInv := rawInvocation{}
		err := json.Unmarshal([]byte(blob), &rawInv)
		assert.Check(t, cmp.ErrorContains(err, "3"), "json.Unmarshal")
	})

	t.Run("null arguments", func(t *testing.T) {
		blob := `["meth", null, "callid"]`
		rawInv := rawInvocation{}
		err := json.Unmarshal([]byte(blob), &rawInv)
		assert.Check(t, cmp.ErrorContains(err, "object"), "json.Unmarshal")
	})
}

func TestMarshalInvocation(t *testing.T) {
	rawInv := rawInvocation{Name: "method", CallID: "foo", Args: []byte(`{}`)}
	blob, err := json.Marshal(rawInv)
	assert.NilError(t, err, "json.Marshal")
	assert.Check(t, cmp.Equal(`["method",{},"foo"]`, string(blob)))

	t.Run("null arguments", func(t *testing.T) {
		rawInv := rawInvocation{Name: "method", CallID: "foo", Args: []byte(`null`)}
		_, err := json.Marshal(rawInv)
		assert.Check(t, cmp.ErrorContains(err, "object"), "json.Marshal")
	})
}

type testArgs struct {
	Argument string `json:"arg"`
}

func unmarshalTestArgs(args json.RawMessage) (interface{}, error) {
	ti := testArgs{}
	err := json.Unmarshal(args, &ti)
	return ti, err
}

func TestMarshalRequest(t *testing.T) {
	req := Request{}
	req.Using = []string{"cap"}
	req.Calls = []Invocation{
		{
			Name:   "NAME",
			CallID: "id",
			Args: testArgs{
				Argument: "foo",
			},
		},
	}

	blob, err := json.Marshal(req)
	assert.NilError(t, err, "json.Marshal")
	assert.Check(t, cmp.Equal(`{"using":["cap"],"methodCalls":[["NAME",{"arg":"foo"},"id"]]}`, string(blob)))
}

func TestUnmarshalRequest(t *testing.T) {
	blob := `{"using":["cap"],"methodCalls":[["NAME",{"arg":"foo"},"id"]]}`
	req := Request{}
	err := req.Unmarshal([]byte(blob), map[string]FuncArgsUnmarshal{"NAME": unmarshalTestArgs})
	assert.NilError(t, err, "req.Unmarshal")
	assert.Check(t, cmp.DeepEqual([]string{"cap"}, req.Using))
	assert.Check(t, cmp.DeepEqual([]Invocation{
		{
			Name:   "NAME",
			CallID: "id",
			Args: testArgs{
				Argument: "foo",
			},
		}}, req.Calls))

	t.Run("unknown method", func(t *testing.T) {
		blob := `{"using":["cap"],"methodCalls":[["unknown",{"arg":"foo"},"id"]]}`
		req := Request{}
		err := req.Unmarshal([]byte(blob), map[string]FuncArgsUnmarshal{"NAME": unmarshalTestArgs})
		assert.Check(t, cmp.ErrorContains(err, ErrUnknownMethod.Error()))
	})
}

func TestMarshalResponse(t *testing.T) {
	resp := Response{}
	resp.SessionState = "state!"
	resp.Responses = []Invocation{
		{
			Name:   "NAME",
			CallID: "id",
			Args: testArgs{
				Argument: "foo",
			},
		},
	}

	blob, err := json.Marshal(resp)
	assert.NilError(t, err, "json.Marshal")
	assert.Check(t, cmp.Equal(`{"sessionState":"state!","methodResponses":[["NAME",{"arg":"foo"},"id"]]}`, string(blob)))

	t.Run("with CreatedIDs", func(t *testing.T) {
		resp := Response{}
		resp.SessionState = "state!"
		resp.Responses = []Invocation{
			{
				Name:   "NAME",
				CallID: "id",
				Args: testArgs{
					Argument: "foo",
				},
			},
		}
		resp.CreatedIDs = map[ID]ID{"abc": "abc"}

		blob, err := json.Marshal(resp)
		assert.NilError(t, err, "json.Marshal")
		assert.Check(t, cmp.Equal(`{"createdIds":{"abc":"abc"},"sessionState":"state!","methodResponses":[["NAME",{"arg":"foo"},"id"]]}`, string(blob)))
	})
	t.Run("with error", func(t *testing.T) {
		resp := Response{}
		resp.SessionState = "state!"
		resp.Responses = []Invocation{
			{
				Name:   "error",
				CallID: "id",
				Args: MethodErrorArgs{
					Type: CodeUnknownMethod,
				},
			},
		}
		blob, err := json.Marshal(resp)
		assert.NilError(t, err, "json.Marshal")
		assert.Check(t, cmp.Equal(`{"sessionState":"state!","methodResponses":[["error",{"type":"unknownMethod"},"id"]]}`, string(blob)))
	})
}

func TestUnmarshalResponse(t *testing.T) {
	blob := `{"sessionState":"state!","methodResponses":[["NAME",{"arg":"foo"},"id"]]}`
	resp := Response{}
	err := resp.Unmarshal([]byte(blob), map[string]FuncArgsUnmarshal{"NAME": unmarshalTestArgs})
	assert.NilError(t, err, "resp.Unmarshal")
	assert.Check(t, cmp.Equal("state!", resp.SessionState))
	assert.Check(t, cmp.DeepEqual([]Invocation{
		{
			Name:   "NAME",
			CallID: "id",
			Args: testArgs{
				Argument: "foo",
			},
		}}, resp.Responses))

	t.Run("with error", func(t *testing.T) {
		blob := `{"sessionState":"state!","methodResponses":[["error",{"type":"unknownMethod"},"id"]]}`
		resp := Response{}
		err := resp.Unmarshal([]byte(blob), map[string]FuncArgsUnmarshal{"NAME": unmarshalTestArgs})
		assert.NilError(t, err, "resp.Unmarshal")
		assert.Check(t, cmp.Equal("state!", resp.SessionState))
		assert.Check(t, cmp.DeepEqual([]Invocation{
			{
				Name:   "error",
				CallID: "id",
				Args: MethodErrorArgs{
					Type: CodeUnknownMethod,
				},
			},
		}, resp.Responses))
	})
}
