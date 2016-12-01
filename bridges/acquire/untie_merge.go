package acquirebridge

import (
	"errors"

	"github.com/svera/sackson-server/bridges/acquire/corporation"
)

func (b *AcquireBridge) untieMerge(clientName string, params untieMergeMessageParams) error {
	if params.CorporationIndex < 0 || params.CorporationIndex > 6 {
		return errors.New(CorporationNotFound)
	}

	corp := b.corporations[params.CorporationIndex]
	if err := b.game.UntieMerge(corp); err != nil {
		return err
	}
	b.history = append(b.history, i18n{
		Key: "game.history.untied_merge",
		Arguments: map[string]string{
			"player":      clientName,
			"corporation": corp.(*corporation.Corporation).Name(),
		},
	})

	return nil
}
