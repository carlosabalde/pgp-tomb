package core

import (
	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/core/query"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

func findPublicKey(alias string) *pgp.PublicKey {
	keys := config.GetPublicKeys()
	if key, found := keys[alias]; found {
		return key
	}
	return nil
}

func parseQuery(queryString string) (result query.Query) {
	if queryString != "" {
		var err error
		result, err = query.Parse(queryString)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"query": queryString,
				"error": err,
			}).Fatal("Failed to parse query!")
		}
	} else {
		result = query.True
	}
	return
}
