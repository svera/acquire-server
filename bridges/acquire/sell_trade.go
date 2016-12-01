package acquirebridge

import (
	"errors"
	"strconv"

	acquireInterfaces "github.com/svera/acquire/interfaces"
	"github.com/svera/sackson-server/bridges/acquire/corporation"
)

func (b *AcquireBridge) sellTrade(clientName string, params sellTradeMessageParams) error {
	var err error
	var corp acquireInterfaces.Corporation

	sell := map[acquireInterfaces.Corporation]int{}
	trade := map[acquireInterfaces.Corporation]int{}

	for corpIndex, operation := range params.CorporationsIndexes {
		index, _ := strconv.Atoi(corpIndex)
		if index < 0 || index > 6 {
			return errors.New(CorporationNotFound)
		}
		corp = b.corporations[index]
		sell[corp] = operation.Sell
		trade[corp] = operation.Trade
	}

	if err = b.game.SellTrade(sell, trade); err != nil {
		return err
	}
	for corp, amount := range sell {
		if amount > 0 {
			b.history = append(b.history, i18n{
				Key: "game.history.sold_stock",
				Arguments: map[string]string{
					"player":      clientName,
					"amount":      strconv.Itoa(amount),
					"corporation": corp.(*corporation.Corporation).Name(),
				},
			})
		}
	}
	for corp, amount := range trade {
		if amount > 0 {
			b.history = append(b.history, i18n{
				Key: "game.history.traded_stock",
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
