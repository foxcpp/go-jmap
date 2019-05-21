package jmap

import (
	"encoding/json"
	"testing"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

func TestUnmarshalInvocation(t *testing.T) {
	blob := `["meth", {"arg1":1}, "callid"]`
	method, callId, args, err := UnmarshalInvocation([]byte(blob))
	assert.NilError(t, err, "UnmarshalInvocation")
	assert.Check(t, cmp.Equal("meth", method))
	assert.Check(t, cmp.Equal("callid", callId))
	assert.Check(t, cmp.DeepEqual(`{"arg1":1}`, string(args)))

	t.Run("missing call ID", func(t *testing.T) {
		blob := `["meth", {"arg1":1}]`
		_, _, _, err := UnmarshalInvocation([]byte(blob))
		assert.Check(t, cmp.ErrorContains(err, "3"), "UnmarshalInvocation")
	})

	t.Run("null arguments", func(t *testing.T) {
		blob := `["meth", null, "callid"]`
		_, _, _, err := UnmarshalInvocation([]byte(blob))
		assert.Check(t, cmp.ErrorContains(err, "object"), "UnmarshalInvocation")
	})
}

func TestMarshalInvocation(t *testing.T) {
	blob, err := MarshalInvocation("method", "foo", []byte(`{}`))
	assert.NilError(t, err, "MarshalInvocation")
	assert.Check(t, cmp.Equal(`["method",{},"foo"]`, string(blob)))

	t.Run("null arguments", func(t *testing.T) {
		_, err := MarshalInvocation("method", "foo", []byte(`null`))
		assert.Check(t, cmp.ErrorContains(err, "object"), "MarshalInvocation")
	})
}

type testInvocation struct {
	NameVal   string `json:"-"`
	CallIDVal string `json:"-"`

	Argument string `json:"arg"`
}

func (ti testInvocation) Name() string {
	return ti.NameVal
}

func (ti testInvocation) CallID() string {
	return ti.CallIDVal
}

func unmarshalTestInvocation(methodName, callId string, args json.RawMessage) (Invocation, error) {
	ti := testInvocation{NameVal: methodName, CallIDVal: callId}
	err := json.Unmarshal(args, &ti)
	return ti, err
}

func TestMarshalRequest(t *testing.T) {
	req := Request{}
	req.Using = []string{"cap"}
	req.Calls = []Invocation{testInvocation{
		NameVal:   "NAME",
		CallIDVal: "id",
		Argument:  "foo",
	}}

	blob, err := json.Marshal(req)
	assert.NilError(t, err, "json.Marshal")
	assert.Check(t, cmp.Equal(`{"using":["cap"],"methodCalls":[["NAME",{"arg":"foo"},"id"]]}`, string(blob)))
}

func TestUnmarshalRequest(t *testing.T) {
	blob := `{"using":["cap"],"methodCalls":[["NAME",{"arg":"foo"},"id"]]}`
	req := Request{}
	err := req.Unmarshal([]byte(blob), map[string]FuncInvocationUnmarshal{"NAME": unmarshalTestInvocation})
	assert.NilError(t, err, "req.Unmarshal")
	assert.Check(t, cmp.DeepEqual([]string{"cap"}, req.Using))
	assert.Check(t, cmp.DeepEqual([]Invocation{testInvocation{
		NameVal:   "NAME",
		CallIDVal: "id",
		Argument:  "foo",
	}}, req.Calls))

	t.Run("unknown method", func(t *testing.T) {
		blob := `{"using":["cap"],"methodCalls":[["unknown",{"arg":"foo"},"id"]]}`
		req := Request{}
		err := req.Unmarshal([]byte(blob), map[string]FuncInvocationUnmarshal{"NAME": unmarshalTestInvocation})
		assert.Check(t, cmp.ErrorContains(err, ErrUnknownMethod.Error()))
	})
}

func TestMarshalResponse(t *testing.T) {
	resp := Response{}
	resp.SessionState = "state!"
	resp.Responses = []Invocation{testInvocation{
		NameVal:   "NAME",
		CallIDVal: "id",
		Argument:  "foo",
	}}

	blob, err := json.Marshal(resp)
	assert.NilError(t, err, "json.Marshal")
	assert.Check(t, cmp.Equal(`{"sessionState":"state!","methodResponses":[["NAME",{"arg":"foo"},"id"]]}`, string(blob)))

	t.Run("with CreatedIDs", func(t *testing.T) {
		resp := Response{}
		resp.SessionState = "state!"
		resp.Responses = []Invocation{testInvocation{
			NameVal:   "NAME",
			CallIDVal: "id",
			Argument:  "foo",
		}}
		resp.CreatedIDs = map[ID]ID{"abc": "abc"}

		blob, err := json.Marshal(resp)
		assert.NilError(t, err, "json.Marshal")
		assert.Check(t, cmp.Equal(`{"createdIds":{"abc":"abc"},"sessionState":"state!","methodResponses":[["NAME",{"arg":"foo"},"id"]]}`, string(blob)))
	})
	t.Run("with error", func(t *testing.T) {
		resp := Response{}
		resp.SessionState = "state!"
		resp.Responses = []Invocation{MethodError{
			Type:        CodeUnknownMethod,
			CallIDValue: "id",
		}}
		blob, err := json.Marshal(resp)
		assert.NilError(t, err, "json.Marshal")
		assert.Check(t, cmp.Equal(`{"sessionState":"state!","methodResponses":[["error",{"type":"unknownMethod"},"id"]]}`, string(blob)))
	})
}

func TestUnmarshalResponse(t *testing.T) {
	blob := `{"sessionState":"state!","methodResponses":[["NAME",{"arg":"foo"},"id"]]}`
	resp := Response{}
	err := resp.Unmarshal([]byte(blob), map[string]FuncInvocationUnmarshal{"NAME": unmarshalTestInvocation})
	assert.NilError(t, err, "resp.Unmarshal")
	assert.Check(t, cmp.Equal("state!", resp.SessionState))
	assert.Check(t, cmp.DeepEqual([]Invocation{testInvocation{
		NameVal:   "NAME",
		CallIDVal: "id",
		Argument:  "foo",
	}}, resp.Responses))

	t.Run("with error", func(t *testing.T) {
		blob := `{"sessionState":"state!","methodResponses":[["error",{"type":"unknownMethod"},"id"]]}`
		resp := Response{}
		err := resp.Unmarshal([]byte(blob), map[string]FuncInvocationUnmarshal{"NAME": unmarshalTestInvocation})
		assert.NilError(t, err, "resp.Unmarshal")
		assert.Check(t, cmp.Equal("state!", resp.SessionState))
		assert.Check(t, cmp.DeepEqual([]Invocation{MethodError{
			Type:        CodeUnknownMethod,
			CallIDValue: "id",
		}}, resp.Responses))
	})
}
