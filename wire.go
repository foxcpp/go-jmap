package jmap

import (
	"encoding/json"
	"errors"
	"fmt"
)

type UnknownMethodError struct {
	MethodName string
}

func (ume UnknownMethodError) Error() string {
	return fmt.Sprintf("jmap: unknown method name: %s", ume.MethodName)
}

type Invocation struct {
	Name   string
	CallID string
	Args   interface{}
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

type rawInvocation struct {
	Name   string
	CallID string
	Args   json.RawMessage
}

// The rawRequest is an intermediate result of request JSON deserialization.
type rawRequest struct {
	request

	RawCalls []rawInvocation `json:"methodCalls"`
}

func (r Request) MarshalJSON() ([]byte, error) {
	raw := rawRequest{}
	raw.Using = r.Using
	raw.Calls = r.Calls
	raw.CreatedIDs = r.CreatedIDs
	raw.RawCalls = make([]rawInvocation, 0, len(r.Calls))
	for _, call := range r.Calls {
		argsBlob, err := json.Marshal(call.Args)
		if err != nil {
			return nil, err
		}
		raw.RawCalls = append(raw.RawCalls, rawInvocation{
			Name:   call.Name,
			CallID: call.CallID,
			Args:   argsBlob,
		})
	}
	return json.Marshal(raw)
}

// Unmarshal deserializes Request object from JSON, calling functions from
// invocationCtors to deserialize Invocation objects. Key is method name.
//
// If error is returned, Request object is not changed.
func (r *Request) Unmarshal(data []byte, argsUnmarshallers map[string]FuncArgsUnmarshal) error {
	raw := rawRequest{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	raw.Calls = make([]Invocation, 0, len(raw.RawCalls))
	for _, rawCall := range raw.RawCalls {
		unmarshal, ok := argsUnmarshallers[rawCall.Name]
		if !ok {
			return UnknownMethodError{MethodName: rawCall.Name}
		}

		args, err := unmarshal(rawCall.Args)
		if err != nil {
			return err
		}

		raw.Calls = append(raw.Calls, Invocation{
			Name:   rawCall.Name,
			CallID: rawCall.CallID,
			Args:   args,
		})
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
	RawResponses []rawInvocation `json:"methodResponses"`
}

func (r Response) MarshalJSON() ([]byte, error) {
	raw := rawResponse{response: response(r)}
	raw.RawResponses = make([]rawInvocation, 0, len(r.Responses))
	for _, response := range r.Responses {
		argsBlob, err := json.Marshal(response.Args)
		if err != nil {
			return nil, err
		}
		raw.RawResponses = append(raw.RawResponses, rawInvocation{
			Name:   response.Name,
			CallID: response.CallID,
			Args:   argsBlob,
		})
	}
	return json.Marshal(raw)
}

// Unmarshal deserializes Response object from JSON, calling functions from
// invocationCtors to deserialize Invocation objects. Key is method name.
// Callback for error decoding is added to invocationCtors implicitly.
//
// If error is returned, Response object is not changed.
func (r *Response) Unmarshal(data []byte, argsUnmarshallers map[string]FuncArgsUnmarshal) error {
	raw := rawResponse{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	raw.Responses = make([]Invocation, 0, len(raw.RawResponses))
	for _, rawResp := range raw.RawResponses {
		var unmarshal FuncArgsUnmarshal
		if rawResp.Name != "error" {
			var ok bool
			unmarshal, ok = argsUnmarshallers[rawResp.Name]
			if !ok {
				return UnknownMethodError{MethodName: rawResp.Name}
			}
		} else {
			unmarshal = UnmarshalMethodErrorArgs
		}

		args, err := unmarshal(rawResp.Args)
		if err != nil {
			return err
		}

		raw.Responses = append(raw.Responses, Invocation{
			Name:   rawResp.Name,
			CallID: rawResp.CallID,
			Args:   args,
		})
	}

	// We will not change r if something goes wrong.
	r.CreatedIDs = raw.CreatedIDs
	r.Responses = raw.Responses
	r.SessionState = raw.SessionState

	return nil
}

func (i *rawInvocation) UnmarshalJSON(data []byte) error {
	var methodName, callId string
	var args json.RawMessage

	// Slice so we can detect invalid size.
	triplet := make([]json.RawMessage, 0, 3)
	if err := json.Unmarshal(data, &triplet); err != nil {
		return err
	}
	if len(triplet) != 3 {
		return errors.New("jmap: malformed Invocation object, need exactly 3 elements")
	}

	if err := json.Unmarshal(triplet[0], &methodName); err != nil {
		return err
	}
	if err := json.Unmarshal(triplet[2], &callId); err != nil {
		return err
	}
	args = triplet[1]

	if args[0] != '{' {
		return errors.New("jmap: malformed Invocation object, arguments must be object")
	}

	i.Name = methodName
	i.CallID = callId
	i.Args = args

	return nil
}

func (i rawInvocation) MarshalJSON() ([]byte, error) {
	if i.Args[0] != '{' {
		return nil, errors.New("jmap: malformed Invocation object, arguments must be object")
	}
	return json.Marshal([3]interface{}{i.Name, i.Args, i.CallID})
}

type FuncArgsUnmarshal func(args json.RawMessage) (interface{}, error)
