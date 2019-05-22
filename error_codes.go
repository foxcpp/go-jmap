package jmap

// Code generated using jmap-errors DO NOT EDIT.

// ProblemPrefix is added to corresponding error code when it is reported
// as a problem type in request-level error.
const ProblemPrefix ErrorCode = "urn:ietf:params:jmap:error:"

type ErrorCode string

const (
	// The accountId does not correspond to a valid account.
	CodeAccountNotFound ErrorCode = "accountNotFound"

	// The accountId given corresponds to a valid account, but the account does
	// not support this method or data type.
	CodeAccountNotSupportedByMethod ErrorCode = "accountNotSupportedByMethod"

	// This method call would modify state in an account that is read-only (as
	// returned on the corresponding Account object in the JMAP Session resource).
	CodeAccountReadOnly ErrorCode = "accountReadOnly"

	// An anchor argument was supplied, but it cannot be found in the results of
	// the query.
	CodeAnchorNotFound ErrorCode = "anchorNotFound"

	// The server forbids duplicates and the record already exists in the target
	// account. An existingId property of type Id MUST be included on the error
	// object with the id of the existing record.
	CodeAlreadyExists ErrorCode = "alreadyExists"

	// The server cannot calculate the changes from the state string given by the
	// client.
	CodeCannotCalculateChanges ErrorCode = "cannotCalculateChanges"

	// The action would violate an ACL or other permissions policy.
	CodeForbidden ErrorCode = "forbidden"

	// The fromAccountId does not correspond to a valid account.
	CodeFromAccountNotFound ErrorCode = "fromAccountNotFound"

	// The fromAccountId given corresponds to a valid account, but the account
	// does not support this data type.
	CodeFromAccountNotSupportedByMethod ErrorCode = "fromAccountNotSupportedByMethod"

	// One of the arguments is of the wrong type or otherwise invalid, or a
	// required argument is missing.
	CodeInvalidArguments ErrorCode = "invalidArguments"

	// The PatchObject given to update the record was not a valid patch.
	CodeInvalidPatch ErrorCode = "invalidPatch"

	// The record given is invalid.
	CodeInvalidProperties ErrorCode = "invalidProperties"

	// The id given cannot be found.
	CodeNotFound ErrorCode = "notFound"

	// The content type of the request was not application/json or the request did
	// not parse as I-JSON.
	CodeNotJSON ErrorCode = "notJSON"

	// The request parsed as JSON but did not match the type signature of the
	// Request object.
	CodeNotRequest ErrorCode = "notRequest"

	// The create would exceed a server-defined limit on the number or total size
	// of objects of this type.
	CodeOverQuota ErrorCode = "overQuota"

	// Too many objects of this type have been created recently, and a
	// server-defined rate limit has been reached. It may work if tried again
	// later.
	CodeRateLimit ErrorCode = "rateLimit"

	// The total number of actions exceeds the maximum number the server is
	// willing to process in a single method call.
	CodeRequestTooLarge ErrorCode = "requestTooLarge"

	// The method used a result reference for one of its arguments, but this
	// failed to resolve.
	CodeInvalidResultReference ErrorCode = "invalidResultReference"

	// An unexpected or unknown error occurred during the processing of the call.
	// The method call made no changes to the server's state.
	CodeServerFail ErrorCode = "serverFail"

	// Some, but not all expected changes described by the method occurred. The
	// client MUST re-synchronise impacted data to determine server state. Use of
	// this error is strongly discouraged.
	CodeServerPartialFail ErrorCode = "serverPartialFail"

	// Some internal server resource was temporarily unavailable. Attempting the
	// same operation later (perhaps after a backoff with a random factor) may
	// succeed.
	CodeServerUnavailable ErrorCode = "serverUnavailable"

	// This is a singleton type, so you cannot create another one or destroy the
	// existing one.
	CodeSingleton ErrorCode = "singleton"

	// An ifInState argument was supplied and it does not match the current state.
	CodeStateMismatch ErrorCode = "stateMismatch"

	// The action would result in an object that exceeds a server-defined limit
	// for the maximum size of a single object of this type.
	CodeTooLarge ErrorCode = "tooLarge"

	// There are more changes than the client's maxChanges argument.
	CodeTooManyChanges ErrorCode = "tooManyChanges"

	// The client included a capability in the "using" property of the request
	// that the server does not support.
	CodeUnknownCapability ErrorCode = "unknownCapability"

	// The server does not recognise this method name.
	CodeUnknownMethod ErrorCode = "unknownMethod"

	// The filter is syntactically valid, but the server cannot process it.
	CodeUnsupportedFilter ErrorCode = "unsupportedFilter"

	// The sort is syntactically valid, but includes a property the server does
	// not support sorting on, or a collation method it does not recognise.
	CodeUnsupportedSort ErrorCode = "unsupportedSort"

	// The client requested an object be both updated and destroyed in the same
	// /set request, and the server has decided to therefore ignore the update.
	CodeWillDestroy ErrorCode = "willDestroy"

	// The mailbox still has at least one child mailbox. The client MUST remove
	// these before it can delete the parent mailbox.
	CodeMailboxHasChild ErrorCode = "mailboxHasChild"

	// The mailbox has at least one message assigned to it and the
	// onDestroyRemoveMessages argument was false.
	CodeMailboxHasEmail ErrorCode = "mailboxHasEmail"

	// At least one blob id referenced in the object doesn't exist.
	CodeBlobNotFound ErrorCode = "blobNotFound"

	// The change to the email's keywords would exceed a server-defined maximum.
	CodeTooManyKeywords ErrorCode = "tooManyKeywords"

	// The change to the email's mailboxes would exceed a server-defined maximum.
	CodeTooManyMailboxes ErrorCode = "tooManyMailboxes"

	// The email to be sent is invalid in some way.
	CodeInvalidEmail ErrorCode = "invalidEmail"

	// The [RFC5321] envelope (supplied or generated) has more recipients than the
	// server allows.
	CodeTooManyRecipients ErrorCode = "tooManyRecipients"

	// The [RFC5321] envelope (supplied or generated) does not have any rcptTo
	// emails.
	CodeNoRecipients ErrorCode = "noRecipients"

	// The rcptTo property of the [RFC5321] envelope (supplied or generated)
	// contains at least one rcptTo value which is not a valid email for sending
	// to.
	CodeInvalidRecipients ErrorCode = "invalidRecipients"

	// The server does not permit the user to send an email with the [RFC5321]
	// envelope From.
	CodeForbiddenMailFrom ErrorCode = "forbiddenMailFrom"

	// The server does not permit the user to send an email with the [RFC5322]
	// From header field of the email to be sent.
	CodeForbiddenFrom ErrorCode = "forbiddenFrom"

	// The user does not have permission to send at all right now.
	CodeForbiddenToSend ErrorCode = "forbiddenToSend"
)
