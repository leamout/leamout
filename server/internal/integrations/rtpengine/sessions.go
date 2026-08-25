package rtpengine

import (
	"fmt"
	"strings"
	"time"
)

func NewSession(callID, fromTag, branch string) (Session, error) {
	session := Session{
		CallID:    strings.TrimSpace(callID),
		FromTag:   strings.TrimSpace(fromTag),
		Branch:    strings.TrimSpace(branch),
		CreatedAt: time.Now().UTC(),
	}
	if err := session.Validate(); err != nil {
		return Session{}, err
	}
	return session, nil
}

func (s Session) WithToTag(toTag string) Session {
	s.ToTag = strings.TrimSpace(toTag)
	return s
}

func (s Session) params(sdp string, flags []string) map[string]any {
	params := map[string]any{
		"call-id":    s.CallID,
		"from-tag":   s.FromTag,
		"via-branch": s.Branch,
	}

	if s.ToTag != "" {
		params["to-tag"] = s.ToTag
	}
	if sdp != "" {
		params["sdp"] = sdp
	}
	if len(flags) > 0 {
		params["flags"] = append([]string(nil), flags...)
	}

	return params
}

func (s Session) String() string {
	if s.ToTag == "" {
		return fmt.Sprintf("%s/%s/%s", s.CallID, s.FromTag, s.Branch)
	}
	return fmt.Sprintf("%s/%s/%s/%s", s.CallID, s.FromTag, s.ToTag, s.Branch)
}
