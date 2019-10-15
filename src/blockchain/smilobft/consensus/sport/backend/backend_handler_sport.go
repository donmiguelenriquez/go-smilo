// Copyright 2019 The go-smilo Authors
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

package backend

import (
	"github.com/ethereum/go-ethereum/common"
	lru "github.com/hashicorp/golang-lru"

	"github.com/ethereum/go-ethereum/log"

	"go-smilo/src/blockchain/smilobft/consensus"
	"go-smilo/src/blockchain/smilobft/consensus/sport"
	"go-smilo/src/blockchain/smilobft/core/types"
	"go-smilo/src/blockchain/smilobft/p2p"
)

// HandleMsg implements consensus.Handler.HandleMsg
func (sb *backend) HandleMsg(addr common.Address, msg p2p.Msg) (bool, error) {
	sb.coreMu.Lock()
	defer sb.coreMu.Unlock()

	if msg.Code == smilobftMsg {
		//log.Debug("HandleMsg, Gossip, smilobftMsg message, GOT IT!!! ", "msg", msg.String())
		if !sb.coreStarted {
			return true, sport.ErrStoppedEngine
		}

		var data []byte
		if err := msg.Decode(&data); err != nil {
			return true, errDecodeFailed
		}

		hash := types.RLPHash(data)

		// Mark peer's message
		ms, ok := sb.recentMessages.Get(addr)
		var m *lru.ARCCache
		if ok {
			m, _ = ms.(*lru.ARCCache)
		} else {
			m, _ = lru.NewARC(inmemoryMessages)
			sb.recentMessages.Add(addr, m)
		}
		m.Add(hash, true)

		// Mark self known message
		if _, ok := sb.knownMessages.Get(hash); ok {
			return true, nil
		}
		sb.knownMessages.Add(hash, true)

		go func() {
			err := sb.smilobftEventMux.Post(sport.MessageEvent{
				Payload: data,
			})
			if err != nil {
				log.Error("Could not send sb.smilobftEventMux.Post, sport.MessageEvent", "err", err)
			} else {
				log.Error("Sent sb.smilobftEventMux.Post, sport.MessageEvent")
			}
		}()

		return true, nil
	}

	return false, nil
}

// SetBroadcaster implements consensus.Handler.SetBroadcaster
func (sb *backend) SetBroadcaster(broadcaster consensus.Broadcaster) {
	sb.broadcaster = broadcaster
}

func (sb *backend) NewChainHead() error {
	sb.coreMu.RLock()
	defer sb.coreMu.RUnlock()
	if !sb.coreStarted {
		return sport.ErrStoppedEngine
	}
	go func() {
		err := sb.smilobftEventMux.Post(sport.FinalCommittedEvent{})
		if err != nil {
			log.Error("NewChainHead, Could not send FinalCommittedEvent, ", "err", err)
		}

	}()
	return nil
}
