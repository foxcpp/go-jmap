package client

import (
	"strconv"

	"github.com/foxcpp/go-jmap"
)

// Batch structure is a helper that makes it easier to construct series of
// method calls to invoke within one request.
type Batch struct {
	req *jmap.Request
}

// NextCallID returns call ID value suitable for use for next request object
// added using Add.
//
// Note that the returned value is changed only by Add call, not by NextCallID itself.
//		a := bt.NextCallID()
//		b := bt.NextCallID()
//		bt.Add(...)
//		c := bt.NextCallID()
//      // a == b, c != b, c != a
func (b *Batch) NextCallID() string {
	return strconv.Itoa(len(b.req.Calls))
}

// NthCallID returns call ID that is or will be used by the N-th call in the
// request.
//
// n is 1-based index, e.g. first request will have n = 1.
func (b *Batch) NthCallID(n int) string {
	return strconv.Itoa(n - 1)
}

// Use adds capability to "using" list of the constructed request.
func (b *Batch) Use(capability string) {
	b.req.Using = append(b.req.Using, capability)
}

// Add adds method call to the constructed request object, using NextCallID for
// call ID value.
func (b *Batch) Add(methodName string, args interface{}) {
	b.req.Calls = append(b.req.Calls, jmap.Invocation{
		Name:   methodName,
		CallID: b.NextCallID(),
		Args:   args,
	})
}

// Request returns Request object constructed by Batch.
//
// Value referenced by pointer should not be changed directly and is valid at
// least until next call to Batch method.
func (b *Batch) Request() *jmap.Request {
	return b.req
}
