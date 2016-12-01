package acquirebridge

import (
	"errors"
	"strconv"

	acquireInterfaces "github.com/svera/acquire/interfaces"
	"github.com/svera/sackson-server/bridges/acquire/corporation"
)

func (b *AcquireBridge) buyStock(clientName string, params buyMessageParams) error {
	buy := map[acquireInterfaces.Corporation]int{}

	for corpIndex, amount := range params.CorporationsIndexes {
		index, _ := strconv.Atoi(corpIndex)
		if index < 0 || index > 6 {
			return errors.New(CorporationNotFound)
		}

		buy[b.corporations[index]] = amount
	}

	if err := b.game.BuyStock(buy); err != nil {
		return err
	}
	for corp, amount := range buy {
		if amount > 0 {
			b.history = append(b.history, i18n{
				Key: "game.history.bought_stock",
				Arguments: map[string]string{
					"player":      clientName,
					"amount":      strconv.Itoa(amount),
					"corporation": corp.(*corporation.Corporation).Name(),
				},
			})
		}
	}
	return nil
}
