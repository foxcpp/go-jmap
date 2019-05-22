// The client package implements base wrappers for JMAP Core protocol client.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/foxcpp/go-jmap"
)

// The Client object wraps *http.Client and stores all information necessary to
// make JMAP API requests.
//
// It is safe to use by multiple goroutines.
type Client struct {
	// HTTPClient to use for requests. If nil - http.DefaultClient is used.
	HTTPClient *http.Client

	// Value of Authentication header.
	Authentication string

	SessionEndpoint string
	BlobEndpoint    string

	// Last seen Session object, set by UpdateSession which is implicitly
	// called on first API request.
	Session *jmap.Session
	// Mutex that is used for access coordination to Session object.
	SessionLck sync.RWMutex
}

func (c *Client) UpdateSession() (*jmap.Session, error) {
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequest("GET", c.SessionEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authentication", c.Authentication)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, decodeError(resp)
	}

	var session jmap.Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, err
	}
	c.SessionLck.Lock()
	c.Session = &session
	c.SessionLck.Unlock()
	return c.Session, nil
}

func (c *Client) lazyInitSession() (jmap.Session, error) {
	c.SessionLck.RLock()
	defer c.SessionLck.RUnlock()
	if c.Session == nil {
		c.SessionLck.RUnlock()
		if _, err := c.UpdateSession(); err != nil {
			return jmap.Session{}, err
		}
		c.SessionLck.RLock()
	}
	return *c.Session, nil
}

func (c *Client) RawSend(r *jmap.Request) (*jmap.Response, error) {
	session, err := c.lazyInitSession()
	if err != nil {
		return nil, err
	}

	if jmap.UnsignedInt(len(r.Calls)) > session.CoreCapability.MaxCallsInRequest {
		return nil, jmap.RequestError{
			Type: jmap.ProblemPrefix + "limit",
			Properties: map[string]interface{}{
				"limit": "maxCallsInRequest",
			},
		}
	}

	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	reqBlob, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", session.APIURL, bytes.NewReader(reqBlob))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authentication", c.Authentication)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, decodeError(resp)
	}

	var response jmap.Response
	return &response, json.NewDecoder(resp.Body).Decode(&response)
}

// Echo sends empty Core/echo request, testing server connectivity.
func (c *Client) Echo() error {
	_, err := c.RawSend(&jmap.Request{Calls: []jmap.Invocation{
		{
			Name:   "Core/echo",
			CallID: "echo0",
			Args:   map[string]interface{}{},
		},
	}})
	return err
}

func decodeError(resp *http.Response) error {
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		return fmt.Errorf("HTTP %d %s", resp.StatusCode, resp.Status)
	}

	var requestErr jmap.RequestError
	if err := json.NewDecoder(resp.Body).Decode(&requestErr); err != nil {
		return fmt.Errorf("HTTP %d %s (failed to decode JSON body: %v)", resp.StatusCode, resp.Status, err)
	}

	return requestErr
}
