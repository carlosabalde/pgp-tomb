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

func (self stringMatch) Eval(context Context) bool {
	return self.regexp.Match([]byte(context.GetIdentifier(self.identifier)))
}

func (self stringMatch) String() string {
	return sprintf("(%s ~ '%s')", self.identifier, self.regexpString)
}

func Match(identifier, regexpString string) (Query, error) {
	regexp, err := regexp.Compile(regexpString)
	if err != nil {
		return nil, errors.Errorf("failed to compile regexp '%s'", regexpString)
	}
	return &stringMatch{identifier, regexpString, regexp}, nil
}
