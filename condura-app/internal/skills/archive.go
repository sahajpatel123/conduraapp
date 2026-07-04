package skills

import (
	"encoding/json"
	"fmt"
)

// ParseArchive unmarshals a hub skill archive (JSON) into a Skill.
func ParseArchive(data []byte) (*Skill, error) {
	var sk Skill
	if err := json.Unmarshal(data, &sk); err != nil {
		return nil, fmt.Errorf("skills: parse archive: %w", err)
	}
	if sk.ID == "" || sk.Name == "" {
		return nil, fmt.Errorf("skills: archive missing id or name")
	}
	return &sk, nil
}

// MarshalArchive serializes a skill for hub publish/download.
func MarshalArchive(sk *Skill) ([]byte, error) {
	if sk == nil {
		return nil, fmt.Errorf("skills: nil skill")
	}
	return json.Marshal(sk)
}
