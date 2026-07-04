package updater

import "errors"

// ErrRestartRequired means the update was staged and needs a process restart
// (always on Windows; recommended on Unix after binary swap).
var ErrRestartRequired = errors.New("updater: restart required to complete update")
