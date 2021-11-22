package journal

import (
	"net/http"
)

func ToJournalHeader(header http.Header) map[string]string {
	journalHeader := make(map[string]string, len(header))
	for key, values := range header {
		journalHeader[key] = values[0]
	}

	return journalHeader
}
