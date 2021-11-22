package journal

import (
	gen "github.com/bluekaki/pkg/id"
)

const JournalHeader = "Journal-Id"

type Journal struct {
	ID          string      `json:"id"`
	Request     *Request    `json:"request"`
	Responses   []*Response `json:"responses"`
	Success     bool        `json:"success"`
	CostSeconds float64     `json:"cost_seconds"`
}

func NewJournal(id string) *Journal {
	if id == "" {
		id = gen.JournalID()
	}

	return &Journal{
		ID: id,
	}
}

func (j *Journal) AppendResponse(resp *Response) {
	if resp != nil {
		j.Responses = append(j.Responses, resp)
	}
}

type Request struct {
	TTL        string            `json:"ttl"`
	Method     string            `json:"method"`
	DecodedURL string            `json:"decoded_url"`
	Header     map[string]string `json:"header"`
	Body       string            `json:"body"`
}

type Response struct {
	Header      map[string]string `json:"header"`
	StatusCode  int               `json:"status_code"`
	Status      string            `json:"status"`
	Body        interface{}       `json:"body"`
	CostSeconds float64           `json:"cost_seconds"`
}
