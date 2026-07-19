package daemon

import (
	"encoding/json"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/ipc"
)

// parseParams unmarshals raw into dest, returning a
// CodeInvalidParams IPC error on failure. Replaces the 4-line
// pattern that appears in 17+ IPC handlers:
//
//	if err := json.Unmarshal(params, &p); err != nil {
//	    return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
//	}
//
// The IPC error code is part of the public JSON-RPC contract;
// pinning it through this helper means every call-site gets
// the same shape automatically.
//
// Note: the error message is json.Unmarshal's verbatim output,
// which includes the byte offset of the parse failure. The CLI
// surfaces this to the user; the GUI displays it in a toast.
func parseParams(raw json.RawMessage, dest any) error {
	if err := json.Unmarshal(raw, dest); err != nil {
		return &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
	}
	return nil
}

// requireField returns a CodeInvalidParams IPC error if value
// is the empty string. Use for path/id/name fields that the
// handler cannot meaningfully operate on without.
//
// Replaces the 3-line pattern that appears in 10+ IPC handlers:
//
//	if p.Path == "" {
//	    return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "path is required"}
//	}
//
// The field name in the error message comes from the caller,
// not from reflection — keeps the helper dependency-free and
// the error message stable (no field-name drift if the struct
// field is renamed).
//
// Future: requireFields(value1, value2, ...) for handlers
// that need multiple required fields with a single error
// message (e.g. "locale is required" vs "code and state are
// required"). Out of scope for this iteration; the per-field
// helper covers the common case.
func requireField(name, value string) error {
	if value == "" {
		return &ipc.Error{Code: ipc.CodeInvalidParams, Message: name + " is required"}
	}
	return nil
}
