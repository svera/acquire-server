package acquirebridge

import "errors"

func (b *AcquireBridge) claimEndGame(clientName string) error {
	if !b.game.ClaimEndGame().IsLastRound() {
		return errors.New(NotEndGame)
	}
	b.history = append(b.history, i18n{
		Key: "game.history.claimed_end",
		Arguments: map[string]string{
			"player": clientName,
		},
	})

	return nil
}
