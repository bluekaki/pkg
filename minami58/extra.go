package minami58

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
)

const (
	descPrefix = "Desc: "
	lineLenght = 64
)

var (
	// ErrDescContainsDelim desc can't contains delim
	ErrDescContainsDelim = errors.New("desc contains delim")

	// Non no delim
	Non Delim = &delim{value: ""}

	// CR \r
	CR Delim = &delim{value: "\r"}

	// LF \n
	LF Delim = &delim{value: "\n"}

	// CRLF \r\n
	CRLF Delim = &delim{value: "\r\n"}
)

var _ Delim = (*delim)(nil)

// Delim delimiter
type Delim interface {
	string() string
	isNon() bool
}

type delim struct {
	value string
}

func (d *delim) string() string {
	return d.value
}

func (d *delim) isNon() bool {
	return d.value == ""
}

func containsDelim(desc string) bool {
	if len(desc) == 0 {
		return false
	}

	return strings.Contains(desc, CR.string()) ||
		strings.Contains(desc, LF.string()) ||
		strings.Contains(desc, CRLF.string())

}

// EncodeWithDesc add some desc into payload
func EncodeWithDesc(raw []byte, desc string, delim Delim) ([]byte, error) {
	if containsDelim(desc) {
		return nil, ErrDescContainsDelim
	}

	buf := bytes.NewBuffer(nil)

	buf.WriteString(descPrefix)
	buf.WriteString(desc)
	if delim.isNon() {
		buf.WriteString(LF.string())
	} else {
		buf.WriteString(delim.string())
	}

	raw = Encode(raw)
	length := len(raw)
	for k := 0; k < length; k += lineLenght {
		if k+lineLenght > length {
			buf.Write(raw[k:])
		} else {
			buf.Write(raw[k : k+lineLenght])
		}

		buf.WriteString(delim.string())
	}

	return buf.Bytes(), nil
}

// DecodeWithDesc decode and return desc
func DecodeWithDesc(raw []byte) (desc string, payload []byte, err error) {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	buf := bytes.NewBuffer(nil)

	firstLine := true
	for scanner.Scan() {
		line := scanner.Bytes()

		if firstLine && strings.HasPrefix(string(line), descPrefix) {
			firstLine = false
			desc = string(line[len(descPrefix):])
			continue
		}

		buf.Write(scanner.Bytes())
	}

	if err = scanner.Err(); err != nil {
		return
	}

	payload, err = Decode(buf.Bytes())
	return
}
