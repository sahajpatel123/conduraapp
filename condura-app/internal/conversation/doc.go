// Package conversation provides SQLite-backed storage for
// chat conversations.
//
// The package name and table name are both "conversations"
// (plural) because a single message lives inside a single
// conversation record. The schema is general enough to
// support either a single in-progress conversation (the
// current Phase 2 behavior) or full history if we want it
// later.
//
// The current API supports two key operations:
//
//   - Create + Append + Get — the GUI's chat view
//     reads and writes through this path. Each Append
//     returns the updated conversation state.
//   - ListRecent — for the chat history sidebar. The
//     "N recent" filter keeps the query O(1) in the
//     number of historical conversations (it just reads
//     the most-recent N from the index, not the whole
//     table).
//
// The package is local-only: no daemon IPC required. The
// store reads/writes SQLite via the same connection pool
// as the rest of the condura daemon (internal/storage).
//
// Adding history support is a schema migration plus a new
// "list conversations since timestamp" query, not a new
// package. The existing API surface stays the same; only
// the storage implementation grows.
package conversation
