package query

import (
	"regexp"

	"github.com/pkg/errors"
)

type stringMatch struct {
	identifier   string
	regexpString string
	regexp       *regexp.Regexp
}

func (m stringMatch) Eval(p Params) bool {
	return m.regexp.Match([]byte(p.Get(m.identifier)))
}

func (m stringMatch) String() string {
	return sprintf("(%s ~ '%s')", m.identifier, m.regexpString)
}

func Match(identifier, regexpString string) (Query, error) {
	regexp, err := regexp.Compile(regexpString)
	if err != nil {
		return nil, errors.Errorf("failed to compile regexp '%s'", regexpString)
	}
	return stringMatch{identifier, regexpString, regexp}, nil
}
