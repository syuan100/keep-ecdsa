package client

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
)

// checkStatusAndRegisterForApplication checks whether the operator is
// registered as a member candidate for keep for the given application.
// If not, checks operators's eligibility and retries until the operator is
// eligible. Eventually, once the operator is eligible, it is registered
// as a keep member candidate.
// Also, once the client is confirmed as registered, it triggers the monitoring
// process to keep the operator's status up to date in the pool.
func checkStatusAndRegisterForApplication(
	ctx context.Context,
	ethereumChain eth.Handle,
	application common.Address,
) {
	isRegistered, err := ethereumChain.IsRegisteredForApplication(application)
	if err != nil {
		logger.Errorf(
			"failed to check if member is registered for application [%s]: [%v]",
			application.String(),
			err,
		)
		return
	}

	go func() {
		if !isRegistered {
			// if the operator is not registered, we need to register it and
			// wait until registration is confirmed
			registerAsMemberCandidate(ctx, ethereumChain, application)
			waitUntilRegistered(ctx, ethereumChain, application)
		}

		// once the registration is confirmed or if the client is already
		// registered, we can start to monitor the status
		monitorSignerPoolStatus(ctx, ethereumChain, application)
	}()
}

// registerAsMemberCandidate checks current operator's eligibility to become
// keep member candidate for the given application and if it is positive,
// registers the operator as a keep member candidate for the given application.
// If the operator is not eligible, it executes the check for each new mined
// block until the operator is finally eligible and can be registered.
func registerAsMemberCandidate(
	parentCtx context.Context,
	ethereumChain eth.Handle,
	application common.Address,
) {
	// If the operator is eligible right now for registering as a member
	// candidate for the application, we register the operator.
	isEligible, err := ethereumChain.IsEligibleForApplication(application)
	if err != nil {
		logger.Errorf(
			"failed to check operator eligibility for application [%s]: [%v]",
			application.String(),
			err,
		)
	}
	if isEligible {
		logger.Infof(
			"registering member candidate for application [%s]",
			application.String(),
		)
		err := ethereumChain.RegisterAsMemberCandidate(application)
		if err != nil {
			logger.Errorf(
				"failed to register member candidate for application [%s]: [%v]",
				application.String(),
				err,
			)
		} else {
			return
		}
	}

	// If the operator is not yet eligible to be registered as a member candidate
	// for the application, we start monitoring eligibility each now block.
	// We do the same in case the registration of eligible operator failed for
	// some reason. As soon as the operator is eligible, we will proceed with
	// the registration.
	registerAsMemberCandidateWhenEligible(parentCtx, ethereumChain, application)
}

// registerAsMemberCandidateWhenEligible for each new block checks the operator's
// eligibility to be registered as a keep member candidate for the application.
// As soon as the operator becomes eligible, function triggers the registration.
func registerAsMemberCandidateWhenEligible(
	parentCtx context.Context,
	ethereumChain eth.Handle,
	application common.Address,
) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	newBlockChan := ethereumChain.BlockCounter().WatchBlocks(ctx)
	for {
		select {
		case <-newBlockChan:
			isEligible, err := ethereumChain.IsEligibleForApplication(application)
			if err != nil {
				logger.Errorf(
					"failed to check operator eligibility for application [%s]: [%v]",
					application.String(),
					err,
				)
				continue
			}

			if !isEligible {
				// if the operator is not yet eligible wait for the next
				// block and execute the check again
				logger.Warningf(
					"operator is not eligible for application [%s]",
					application.String(),
				)
				continue
			}

			// if the operator is eligible, register it as a keep member
			// candidate for this application
			logger.Infof(
				"registering member candidate for application [%s]",
				application.String(),
			)
			if err := ethereumChain.RegisterAsMemberCandidate(application); err != nil {
				logger.Errorf(
					"failed to register member candidate for application [%s]: [%v]",
					application.String(),
					err,
				)
				continue
			}

			// we cancel the context in case the registration was successful,
			// we don't want to do it again
			cancel()
		case <-ctx.Done():
			return
		}
	}
}

// waitUntilRegistered blocks until the operator is registered as a keep member
// candidate for the given application.
func waitUntilRegistered(
	ctx context.Context,
	ethereumChain eth.Handle,
	application common.Address,
) {
	newBlockChan := ethereumChain.BlockCounter().WatchBlocks(ctx)

	for {
		select {
		case <-newBlockChan:
			isRegistered, err := ethereumChain.IsRegisteredForApplication(application)
			if err != nil {
				logger.Errorf(
					"failed to check if member is registered for application [%s]: [%v]",
					application.String(),
					err,
				)
				continue
			}

			if isRegistered {
				logger.Infof(
					"operator is registered for application [%s]",
					application.String(),
				)
				return
			}

			logger.Infof(
				"operator is not yet registered for application [%s], waiting...",
				application.String(),
			)
		case <-ctx.Done():
			return
		}
	}
}

// monitorSignerPoolStatus tracks operator's state in the signing pool
// (staking weight, bonding) and updates the status when it gets out of date.
func monitorSignerPoolStatus(
	ctx context.Context,
	ethereumChain eth.Handle,
	application common.Address,
) {
	logger.Debugf(
		"starting monitoring operatator status for application [%s]",
		application.String(),
	)

	newBlockChan := ethereumChain.BlockCounter().WatchBlocks(ctx)

	for {
		select {
		case <-newBlockChan:
			isUpToDate, err := ethereumChain.IsStatusUpToDateForApplication(application)
			if err != nil {
				logger.Errorf(
					"failed to check operator status for application [%s]: [%v]",
					application.String(),
					err,
				)
				continue
			}

			if isUpToDate {
				logger.Debugf(
					"operator status is up to date for application [%s]",
					application.String(),
				)
			} else {
				logger.Infof(
					"updating operator status for application [%s]",
					application.String(),
				)

				err := ethereumChain.UpdateStatusForApplication(application)
				if err != nil {
					logger.Errorf(
						"failed to update operator status for application [%s]: [%v]",
						application.String(),
						err,
					)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}