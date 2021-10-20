package proposal

import (
	"encoding/json"
	"time"

	"github.com/bluekaki/pkg/errors"
)

type AlertMessage struct {
	ProjectName  string            `json:"project_name,omitempty"`
	JournalID    string            `json:"journal_id,omitempty"`
	ErrorVerbose string            `json:"error_verbose,omitempty"`
	Meta         *AlertMessageMeta `json:"meta,omitempty"`
	Timestamp    time.Time         `json:"timestamp,omitempty"`
}

type AlertMessageMeta struct {
	URL        string `json:"url,omitempty"`
	Parameters string `json:"parameters,omitempty"`
}

func (a *AlertMessage) Marshal() []byte {
	raw, _ := json.Marshal(a)
	return raw
}

func (a *AlertMessage) Unmarshal(raw []byte) error {
	if err := json.Unmarshal(raw, a); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
