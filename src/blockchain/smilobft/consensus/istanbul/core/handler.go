// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"github.com/ethereum/go-ethereum/common"
	"go-smilo/src/blockchain/smilobft/consensus/istanbul"
)

// Start implements core.Engine.Start
func (c *core) Start() error {
	// Start a new round from last sequence + 1
	c.startNewRound(common.Big0)

	// Tests will handle events itself, so we have to make subscribeEvents()
	// be able to call in test.
	c.subscribeEvents()
	go c.handleEvents()

	return nil
}

// Stop implements core.Engine.Stop
func (c *core) Stop() error {
	c.stopTimer()
	c.unsubscribeEvents()

	// Make sure the handler goroutine exits
	c.handlerStopCh <- struct{}{}
	return nil
}

// ----------------------------------------------------------------------------

// Subscribe both internal and external events
func (c *core) subscribeEvents() {
	c.events = c.backend.EventMux().Subscribe(
		// external events
		istanbul.RequestEvent{},
		istanbul.MessageEvent{},
		// internal events
		backlogEvent{},
	)
	c.timeoutSub = c.backend.EventMux().Subscribe(
		timeoutEvent{},
	)
	c.finalCommittedSub = c.backend.EventMux().Subscribe(
		istanbul.FinalCommittedEvent{},
	)
}

// Unsubscribe all events
func (c *core) unsubscribeEvents() {
	c.events.Unsubscribe()
	c.timeoutSub.Unsubscribe()
	c.finalCommittedSub.Unsubscribe()
}

func (c *core) handleEvents() {
	// Clear state
	defer func() {
		c.current = nil
		<-c.handlerStopCh
	}()

	for {
		select {
		case event, ok := <-c.events.Chan():
			if !ok {
				return
			}
			// A real event arrived, process interesting content
			switch ev := event.Data.(type) {
			case istanbul.RequestEvent:
				r := &istanbul.Request{
					Proposal: ev.Proposal,
				}
				err := c.handleRequest(r)
				if err == errFutureMessage {
					c.logger.Debug("$$$ istanbul, handleEvents, RequestEvent arrived, errFutureMessage", "Proposal", ev.Proposal.Hash().Hex())
					c.storeRequestMsg(r)
				}
			case istanbul.MessageEvent:
				if err := c.handleMsg(ev.Payload); err == nil {
					c.logger.Debug("$$$ istanbul, handleEvents, MessageEvent arrived, will send Gossip to fullnodeSet")
					err = c.backend.Gossip(c.valSet, ev.Payload)
					if err != nil {
						c.logger.Error("$$$ istanbul, handleEvents, handleMsg, failed to backend.Gossip", "err", err)
					}
				}else {
					c.logger.Error("$$$ istanbul, handleEvents, handleMsg", "err", err)
				}
			case backlogEvent:
				// No need to check signature for internal messages
				if err := c.handleCheckedMsg(ev.msg, ev.src); err == nil {
					p, err := ev.msg.Payload()
					if err != nil {
						c.logger.Warn("handleEvents, Get message payload failed", "err", err)
						continue
					}
					err = c.backend.Gossip(c.valSet, p)
					if err != nil {
						c.logger.Error("$$$ istanbul, handleEvents, handleCheckedMsg, backend.Gossip ", "err", err)
					}
				}
			}
		case _, ok := <-c.timeoutSub.Chan():
			if !ok {
				return
			}
			c.handleTimeoutMsg()
		case event, ok := <-c.finalCommittedSub.Chan():
			if !ok {
				return
			}
			switch event.Data.(type) {
			case istanbul.FinalCommittedEvent:
				err := c.handleFinalCommitted()
				if err != nil {
					c.logger.Error("$$$ istanbul, handleEvents, FinalCommittedEvent, handleFinalCommitted", "err", err)
				}
			}
		}
	}
}

// sendEvent sends events to mux
func (c *core) sendEvent(ev interface{}) {
	err := c.backend.EventMux().Post(ev)
	if err != nil {
		c.logger.Error("$$$ istanbul, sendEvent", "err", err)
	}
}

func (c *core) handleMsg(payload []byte) error {
	logger := c.logger.New()

	// Decode message and check its signature
	msg := new(message)
	if err := msg.FromPayload(payload, c.validateFn); err != nil {
		logger.Error("Failed to decode message from payload", "err", err)
		return err
	}

	// Only accept message if the address is valid
	_, src := c.valSet.GetByAddress(msg.Address)
	if src == nil {
		logger.Error("Invalid address in message", "msg", msg)
		return istanbul.ErrUnauthorizedAddress
	}

	return c.handleCheckedMsg(msg, src)
}

func (c *core) handleCheckedMsg(msg *message, src istanbul.Validator) error {
	logger := c.logger.New("address", c.address, "from", src)

	// Store the message if it's a future message
	testBacklog := func(err error) error {
		if err == errFutureMessage {
			c.storeBacklog(msg, src)
		}

		return err
	}

	switch msg.Code {
	case msgPreprepare:
		return testBacklog(c.handlePreprepare(msg, src))
	case msgPrepare:
		return testBacklog(c.handlePrepare(msg, src))
	case msgCommit:
		return testBacklog(c.handleCommit(msg, src))
	case msgRoundChange:
		return testBacklog(c.handleRoundChange(msg, src))
	default:
		logger.Error("Invalid message", "msg", msg)
	}

	return errInvalidMessage
}

func (c *core) handleTimeoutMsg() {
	// If we're not waiting for round change yet, we can try to catch up
	// the max round with F+1 round change message. We only need to catch up
	// if the max round is larger than current round.
	if !c.waitingForRoundChange {
		maxRound := c.roundChangeSet.MaxRound(c.valSet.F() + 1)
		if maxRound != nil && maxRound.Cmp(c.current.Round()) > 0 {
			c.sendRoundChange(maxRound)
			return
		}
	}

	lastProposal, _ := c.backend.LastProposal()
	if lastProposal != nil && lastProposal.Number().Cmp(c.current.Sequence()) >= 0 {
		c.logger.Trace("round change timeout, catch up latest sequence", "number", lastProposal.Number().Uint64())
		c.startNewRound(common.Big0)
	} else {
		c.sendNextRoundChange()
	}
}