package jmap

import (
	"encoding/json"
	"errors"
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
	Type ErrorCode `json:"type"`

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

// The MethodError structure describes method-level error.
//
// See section 3.5.2 of JMAP Core specification.
type MethodErrorArgs struct {
	Type ErrorCode `json:"type"`

	// All fields other than Type.
	Properties map[string]interface{}
}

func (me MethodErrorArgs) Error() string {
	return "jmap: " + string(me.Type)
}

func (me MethodErrorArgs) MarshalJSON() ([]byte, error) {
	fullProps := make(map[string]interface{}, len(me.Properties)+1)
	for k, v := range me.Properties {
		fullProps[k] = v
	}
	fullProps["type"] = me.Type

	return json.Marshal(fullProps)
}

func (me *MethodErrorArgs) UnmarshalJSONArgs(data []byte) error {
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

func UnmarshalMethodErrorArgs(args json.RawMessage) (interface{}, error) {
	res := MethodErrorArgs{}
	err := res.UnmarshalJSONArgs(args)
	return res, err
}
