package jmap

import (
	"encoding/json"
	"errors"
)

type ProblemType string

const (
	ProblemUnknownCapability ProblemType = "urn:ietf:params:jmap:error:unknownCapability"
	ProblemNotJSON           ProblemType = "urn:ietf:params:jmap:error:notJSON"
	ProblemNotRequest        ProblemType = "urn:ietf:params:jmap:error:notRequest"
	ProblemLimit             ProblemType = "urn:ietf:params:jmap:error:limit"
)

// The RequestError structure is "problem details" object as defined by
// RFC 7807. Any fields except for Type can be empty.
//
// Used for all problem types that don't have any custom fields. Currently:
// - urn:ietf:params:jmap:error:unknownCapability (ProblemUnknownCapability)
// - urn:ietf:params:jmap:error:notJSON (ProblemNotJSON)
// - urn:ietf:params:jmap:error:notRequest (ProblemNotRequest)
type RequestError struct {
	// A URI reference that identifies the problem type.
	Type ProblemType `json:"type"`

	// A short, human-readable summary of the problem type.
	Title string `json:"title,omitempty"`

	// The HTTP status code.
	Status int `json:"status,omitempty"`

	// A human-readable explanation specific to this occurrence of the problem.
	Detail string `json:"detail,omitempty"`

	// A URI reference that identifies the specific occurrence of the problem.
	Instance string `json:"instance,omitempty"`

	// All other fields.
	Properties map[string]interface{} `json:"-"`
}

func (re RequestError) Error() string {
	return re.Detail
}

type requestError RequestError

func (re *RequestError) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*requestError)(re)); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &re.Properties); err != nil {
		return err
	}
	if _, ok := re.Properties["type"]; !ok {
		return errors.New("jmap: missing type field in error object")
	}
	delete(re.Properties, "type")
	delete(re.Properties, "title")
	delete(re.Properties, "status")
	delete(re.Properties, "detail")
	delete(re.Properties, "instance")
	if len(re.Properties) == 0 {
		re.Properties = nil
	}
	return nil
}

func (re RequestError) MarshalJSON() ([]byte, error) {
	allProps := make(map[string]interface{}, len(re.Properties)+5)
	for k, v := range re.Properties {
		allProps[k] = v
	}

	if re.Type != "" {
		allProps["type"] = re.Type
	}
	if re.Title != "" {
		allProps["title"] = re.Title
	}
	if re.Status != 0 {
		allProps["status"] = re.Status
	}
	if re.Detail != "" {
		allProps["detail"] = re.Detail
	}
	if re.Instance != "" {
		allProps["instance"] = re.Instance
	}
	return json.Marshal(allProps)
}

type MethodErrorType string

const (
	// Some internal server resource was temporarily unavailable. Attempting
	// the same operation later (perhaps after a backoff with a random factor)
	// may succeed.
	MethodErrServUnavailable MethodErrorType = "serverUnavailable"

	// An unexpected or unknown error occurred during the processing of the
	// call. A description property should provide more details about the
	// error. The method call made no changes to the server’s state.
	// Attempting the same operation again is expected to fail again.
	MethodErrServFail = "serverFail"

	// Some, but not all expected changes described by the method occurred.
	// The client MUST re-synchronise impacted data to determine server state.
	// Use of this error is strongly discouraged.
	MethodErrPartialFail = "serverPartialFail"

	// The server does not recognise this method name.
	MethodErrUnknownMethod = "unknownMethod"

	// One of the arguments is of the wrong type or otherwise invalid, or a
	// required argument is missing. A description property MAY be present to
	// help debug with an explanation of what the problem was.
	MethodErrInvalidArgs = "invalidArguments"

	// The method used a result reference for one of its arguments, but this
	// failed to resolve.
	MethodErrInvalidReference = "invalidResultReference"

	// The method and arguments are valid, but executing the method would
	// violate an ACL or other permissions policy.
	MethodErrForbidden = "forbidden"

	// The accountId does not correspond to a valid account.
	MethodErrAcctNotFound = "accountNotFound"

	// The accountId given corresponds to a valid account, but the account does
	// not support this method or data type.
	MethodErrAcctUnsupportedMethod = "accountNotSupportedByMethod"

	// This method call would modify state in an account that is read-only (as
	// returned on the corresponding Account object in the JMAP Session
	// resource).
	MethodErrAcctReadOnly = "accountReadOnly"
)

// The MethodError structure describes method-level error.
//
// See section 3.5.2 of JMAP Core specification.
type MethodError struct {
	Type        MethodErrorType `json:"type"`
	CallIDValue string

	// All fields other than Type.
	Properties map[string]interface{}
}

func (me MethodError) Name() string {
	return "error"
}
func (me MethodError) CallID() string {
	return me.CallIDValue
}

func (me MethodError) Error() string {
	return "jmap: " + string(me.Type)
}

func (me MethodError) MarshalJSON() ([]byte, error) {
	fullProps := make(map[string]interface{}, len(me.Properties)+1)
	for k, v := range me.Properties {
		fullProps[k] = v
	}
	fullProps["type"] = me.Type

	return json.Marshal(fullProps)
}

func (me *MethodError) UnmarshalJSONArgs(data []byte) error {
	if err := json.Unmarshal(data, &me); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &me.Properties); err != nil {
		return err
	}
	if _, ok := me.Properties["type"]; !ok {
		return errors.New("jmap: missing type field in error object")
	}
	delete(me.Properties, "type")
	if len(me.Properties) == 0 {
		me.Properties = nil
	}
	return nil
}

func UnmarshalMethodError(methodName, callId string, args json.RawMessage) (Invocation, error) {
	res := MethodError{CallIDValue: callId}
	err := res.UnmarshalJSONArgs(args)
	return res, err
}
