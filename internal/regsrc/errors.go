package regsrc

import (
	"errors"
	"fmt"
)

type ParserError struct {
	Summary string
	Detail  string
}

func (pe *ParserError) Error() string {
	return fmt.Sprintf("%s: %s", pe.Summary, pe.Detail)
}

var ErrParseProvider = errors.New("error parsing provider")
