package calls

import (
	"context"
	"fmt"
)

// Reconcile only from positive channel evidence. Absence alone cannot prove
// that a queued bgapi originate never ran, so no automatic redial or release.
func (j *ReconciliationJob) reconcileOriginations(ctx context.Context) error {
	attempts, err := j.repo.ListUnresolvedOriginations(ctx)
	if err != nil {
		return err
	}
	for _, attempt := range attempts {
		channelID := attempt.ChannelID
		// A disconnected FreeSWITCH client cannot establish channel absence.
		exists, err := j.service.controller.ChannelExists(ctx, channelID)
		if err != nil {
			return fmt.Errorf("query originate channel %s: %w", channelID, err)
		}
		if !exists {
			continue
		}
		if err := j.service.channels.Bind(ctx, attempt.CallID, channelID); err != nil {
			return fmt.Errorf("bind recovered channel %s: %w", channelID, err)
		}
		if err := j.repo.MarkOriginationObserved(ctx, attempt.CallID, channelID); err != nil {
			return fmt.Errorf("mark recovered channel %s: %w", channelID, err)
		}
	}
	return nil
}

// Successful channel discovery does not imply the call was answered; only
// lifecycle events may advance the call state.
