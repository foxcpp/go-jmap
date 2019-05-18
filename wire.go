package jmap

import (
	"encoding/json"
	"errors"
)

var ErrUnknownMethod = errors.New("jmap: unknown method name")

// Invocation interface describes common subset of operations that are valid on
// any method call response.
//
// json.Marshal on the Invocation should produce a JSON object containing only
// Invocation arguments.
type Invocation interface {
	Name() string
	CallID() string
}

type Request struct {
	// The set of capabilities the client wishes to use. The client may include
	// capability identifiers even if the method calls it makes do not utilise
	// those capabilities.
	Using []string `json:"using"`

	// An array of method calls to process on the server. The method calls will
	// be processed sequentially, in order.
	Calls []Invocation `json:"-"`

	// A map of (client-specified) creation id to the id the server assigned
	// when a record was successfully created. Can be nil.
	CreatedIDs map[ID]ID `json:"createdIds,omitempty"`
}

// The request type is defined to trick encoding/json to not call
// Request.MarshalJSON when marshalling rawRequest.
// This allows us to use struct tags and custom marshalling logic for Calls at
// the same time.
type request Request

// The rawRequest is an intermediate result of request JSON deserialization.
type rawRequest struct {
	request

	RawCalls []json.RawMessage `json:"methodCalls"`
}

func (r Request) MarshalJSON() ([]byte, error) {
	raw := rawRequest{}
	raw.Using = r.Using
	raw.Calls = r.Calls
	raw.CreatedIDs = r.CreatedIDs
	raw.RawCalls = make([]json.RawMessage, 0, len(r.Calls))
	for _, call := range r.Calls {
		argsBlob, err := json.Marshal(call)
		if err != nil {
			return nil, err
		}
		callBlob, err := MarshalInvocation(call.Name(), call.CallID(), argsBlob)
		if err != nil {
			return nil, err
		}

		raw.RawCalls = append(raw.RawCalls, json.RawMessage(callBlob))
	}
	return json.Marshal(raw)
}

// Unmarshal deserializes Request object from JSON, calling functions from
// invocationCtors to deserialize Invocation objects. Key is method name.
//
// If error is returned, Request object is not changed.
func (r *Request) Unmarshal(data []byte, invocationCtors map[string]FuncInvocationUnmarshal) error {
	raw := rawRequest{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	raw.Calls = make([]Invocation, 0, len(raw.RawCalls))
	for _, rawCall := range raw.RawCalls {
		method, callId, args, err := UnmarshalInvocation(rawCall)
		if err != nil {
			return err
		}

		ctor, ok := invocationCtors[method]
		if !ok {
			return ErrUnknownMethod
		}

		invocation, err := ctor(method, callId, args)
		if err != nil {
			return err
		}

		raw.Calls = append(raw.Calls, invocation)
	}

	// We will not change r if something goes wrong.
	r.Using = raw.Using
	r.Calls = raw.Calls
	r.CreatedIDs = raw.CreatedIDs

	return nil
}

type Response struct {
	// An array of responses, in the same format as the Calls on the
	// Request object. The output of the methods will be added to the
	// methodResponses array in the same order as the methods are processed.
	Responses []Invocation `json:"-"`

	// A map of (client-specified) creation id to the id the server assigned
	// when a record was successfully created.
	CreatedIDs map[ID]ID `json:"createdIds,omitempty"`

	// The current value of the “state” string on the JMAP Session object, as
	// described in section 2. Clients may use this to detect if this object
	// has changed and needs to be refetched.
	SessionState string `json:"sessionState"`
}

type response Response

type rawResponse struct {
	response
	RawResponses []json.RawMessage `json:"methodResponses"`
}

func (r Response) MarshalJSON() ([]byte, error) {
	raw := rawResponse{response: response(r)}
	raw.RawResponses = make([]json.RawMessage, 0, len(r.Responses))
	for _, response := range r.Responses {
		argsBlob, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}
		respBlob, err := MarshalInvocation(response.Name(), response.CallID(), argsBlob)
		if err != nil {
			return nil, err
		}

		raw.RawResponses = append(raw.RawResponses, json.RawMessage(respBlob))
	}
	return json.Marshal(raw)
}

// Unmarshal deserializes Response object from JSON, calling functions from
// invocationCtors to deserialize Invocation objects. Key is method name.
// Callback for error decoding is added to invocationCtors implicitly.
//
// If error is returned, Response object is not changed.
func (r *Response) Unmarshal(data []byte, invocationCtors map[string]FuncInvocationUnmarshal) error {
	raw := rawResponse{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	raw.Responses = make([]Invocation, 0, len(raw.RawResponses))
	for _, rawResp := range raw.RawResponses {
		method, callId, args, err := UnmarshalInvocation(rawResp)
		if err != nil {
			return err
		}

		var ctor FuncInvocationUnmarshal
		if method == "error" {
			// We don't change invocationCtors argument because it could affect
			// users code.
			ctor = UnmarshalMethodError
		} else {
			var ok bool
			ctor, ok = invocationCtors[method]
			if !ok {
				return ErrUnknownMethod
			}
		}

		invocation, err := ctor(method, callId, args)
		if err != nil {
			return err
		}

		raw.Responses = append(raw.Responses, invocation)
	}

	// We will not change r if something goes wrong.
	r.CreatedIDs = raw.CreatedIDs
	r.Responses = raw.Responses
	r.SessionState = raw.SessionState

	return nil
}

// UnmarshalInvocation a helper function that unpacks Invocation triplet (name,
// args, callId) into strings for name and callId, returning arguments as raw JSON that can just
// then unmarshalled into actual response structure.
//
// UnmarshalInvocation is a low-level protocol function exported for the use of
// protocols built on JMAP Core and extensions. You probably don't want to use
// it directly.
func UnmarshalInvocation(data []byte) (methodName, callId string, args json.RawMessage, err error) {
	// Slice so we can detect invalid size.
	triplet := make([]json.RawMessage, 0, 3)
	if err := json.Unmarshal(data, &triplet); err != nil {
		return "", "", nil, err
	}
	if len(triplet) != 3 {
		return "", "", nil, errors.New("jmap: malformed Invocation object, need exactly 3 elements")
	}

	if err := json.Unmarshal(triplet[0], &methodName); err != nil {
		return "", "", nil, err
	}
	if err := json.Unmarshal(triplet[2], &callId); err != nil {
		return "", "", nil, err
	}
	args = triplet[1]

	if args[0] != '{' {
		return "", "", nil, errors.New("jmap: malformed Invocation object, arguments must be object")
	}

	return
}

// MarshalInvocation a helper function that packs Invocation triplet (name,
// args, callId) into JSON representation.
//
// args must be non-null JSON object, as required by schema.
//
// MarshalInvocation is a low-level protocol function exported for the use of
// protocols built on JMAP Core and extensions. You probably don't want to use
// it directly.
func MarshalInvocation(methodName, callId string, args json.RawMessage) ([]byte, error) {
	if args[0] != '{' {
		return nil, errors.New("jmap: malformed Invocation object, arguments must be object")
	}
	return json.Marshal([3]interface{}{methodName, args, callId})
}

type FuncInvocationUnmarshal func(methodName, callId string, args json.RawMessage) (Invocation, error)
