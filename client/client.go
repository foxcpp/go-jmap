// The client package implements base wrappers for JMAP Core protocol client.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	// Session endpoint URL. Must be set before any request.
	SessionEndpoint string

	// Last seen Session object, set by UpdateSession which is implicitly
	// called on first API request.
	Session *jmap.Session
	// Mutex that is used for access coordination to Session object.
	SessionLck sync.RWMutex
}

// UpdateSession sets c.Session and returns it.
//
// Session object contains information necessary to do almost all requests so
// UpdateSession is called implicitly on first API request.
func (c *Client) UpdateSession() (*jmap.Session, error) {
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	if c.SessionEndpoint == "" {
		return nil, fmt.Errorf("jmap/client: SessionEndpoint is empty")
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

// RawSend sends manually constructed jmap.Request object and returns parsed
// jmap.Response object.
//
// It initializes c.Session if it is empty.
func (c *Client) RawSend(r *jmap.Request) (*jmap.Response, error) {
	if c.SessionEndpoint == "" {
		return nil, fmt.Errorf("jmap/client: SessionEndpoint is empty")
	}

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

// Upload sends binary data to the server and returns blob ID and some
// associated meta-data.
//
// There are some caveats to keep in mind:
// - Server may return the same blob ID for multiple uploads of the same blob.
// - Blob ID may become invalid after some time if it is unused.
// - Blob ID is usable only by the uploader until it is used, even for shared accounts.
func (c *Client) Upload(account jmap.ID, blob io.Reader) (*jmap.BlobInfo, error) {
	if c.SessionEndpoint == "" {
		return nil, fmt.Errorf("jmap/client: SessionEndpoint is empty")
	}

	session, err := c.lazyInitSession()
	if err != nil {
		return nil, err
	}

	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	tgtUrl := strings.Replace(session.UploadURL, "{accountId}", string(account), -1)
	req, err := http.NewRequest("POST", tgtUrl, blob)
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

	var info jmap.BlobInfo
	return &info, json.NewDecoder(resp.Body).Decode(&info)
}

func (c *Client) Download(account, blob jmap.ID) (io.ReadCloser, error) {
	if c.SessionEndpoint == "" {
		return nil, fmt.Errorf("jmap/client: SessionEndpoint is empty")
	}

	session, err := c.lazyInitSession()
	if err != nil {
		return nil, err
	}

	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	urlRepl := strings.NewReplacer(
		"{accountId}", string(account),
		"{blobId}", string(blob),
		"{type}", "application/octet-stream", // are any other values necessary?
		"{name}", "filename",
	)
	tgtUrl := urlRepl.Replace(session.DownloadURL)
	req, err := http.NewRequest("GET", tgtUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authentication", c.Authentication)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		defer resp.Body.Close()
		return nil, decodeError(resp)
	}

	return resp.Body, nil
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
