// Package autonomy provides per-task-type + per-app autonomy levels.
//
//nolint:revive // enum values are self-documenting
package autonomy

// Level is the autonomy level for a task-app pair.
type Level int

const (
	Unset      Level = -1
	Block      Level = iota
	Warn       Level = iota
	Ask        Level = iota
	Autonomous Level = iota
)

func (l Level) String() string {
	switch l {
	case Block:
		return "block"
	case Warn:
		return "warn"
	case Ask:
		return "ask"
	case Autonomous:
		return "autonomous"
	default:
		return "warn"
	}
}

// Matrix maps (task_type, app) pairs to autonomy levels.
type Matrix struct {
	defaultLevel Level
	mapping      map[string]Level
}

// NewMatrix creates an autonomy matrix from a config map.
func NewMatrix(defaultLevel Level, mapping map[string]Level) *Matrix {
	if defaultLevel == Unset {
		defaultLevel = Warn
	}
	return &Matrix{defaultLevel: defaultLevel, mapping: mapping}
}

// Evaluate returns the autonomy level for a task-app pair.
func (m *Matrix) Evaluate(taskType, app string) Level {
	key := taskType + "." + app
	if lvl, ok := m.mapping[key]; ok {
		return lvl
	}
	// Try task wildcard.
	if lvl, ok := m.mapping[taskType+".*"]; ok {
		return lvl
	}
	return m.defaultLevel
}

// CanAutoApply returns true if the action can be executed without
// consent at this autonomy level. The DESTRUCTIVE carve-out is
// enforced by the caller — autonomous only applies to READ/WRITE.
func CanAutoApply(level Level, isDestructive bool) bool {
	if isDestructive {
		return false
	}
	return level == Autonomous
}

// NeedsConsent returns the consent type required for this level.
func NeedsConsent(level Level, isDestructive bool) (needsConsent, needsPresence bool) {
	if isDestructive {
		return true, true
	}
	switch level {
	case Block:
		return true, true
	case Warn:
		return true, false
	case Ask:
		return true, true
	case Autonomous:
		return false, false
	default:
		return true, false
	}
}
