package consensus

import (
	types "go-smilo/src/blockchain/smilobft/core/types"
	reflect "reflect"

	"github.com/ethereum/go-ethereum/common"
	gomock "github.com/golang/mock/gomock"
)

// MockBroadcaster is a mock of Broadcaster interface
type MockBroadcaster struct {
	ctrl     *gomock.Controller
	recorder *MockBroadcasterMockRecorder
}

// MockBroadcasterMockRecorder is the mock recorder for MockBroadcaster
type MockBroadcasterMockRecorder struct {
	mock *MockBroadcaster
}

// NewMockBroadcaster creates a new mock instance
func NewMockBroadcaster(ctrl *gomock.Controller) *MockBroadcaster {
	mock := &MockBroadcaster{ctrl: ctrl}
	mock.recorder = &MockBroadcasterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBroadcaster) EXPECT() *MockBroadcasterMockRecorder {
	return m.recorder
}

// Enqueue mocks base method
func (m *MockBroadcaster) Enqueue(id string, block *types.Block) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Enqueue", id, block)
}

// Enqueue indicates an expected call of Enqueue
func (mr *MockBroadcasterMockRecorder) Enqueue(id, block interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enqueue", reflect.TypeOf((*MockBroadcaster)(nil).Enqueue), id, block)
}

// FindPeers mocks base method
func (m *MockBroadcaster) FindPeers(arg0 map[common.Address]struct{}) map[common.Address]Peer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindPeers", arg0)
	ret0, _ := ret[0].(map[common.Address]Peer)
	return ret0
}

// FindPeers indicates an expected call of FindPeers
func (mr *MockBroadcasterMockRecorder) FindPeers(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindPeers", reflect.TypeOf((*MockBroadcaster)(nil).FindPeers), arg0)
}

// MockPeer is a mock of Peer interface
type MockPeer struct {
	ctrl     *gomock.Controller
	recorder *MockPeerMockRecorder
}

// MockPeerMockRecorder is the mock recorder for MockPeer
type MockPeerMockRecorder struct {
	mock *MockPeer
}

// NewMockPeer creates a new mock instance
func NewMockPeer(ctrl *gomock.Controller) *MockPeer {
	mock := &MockPeer{ctrl: ctrl}
	mock.recorder = &MockPeerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPeer) EXPECT() *MockPeerMockRecorder {
	return m.recorder
}

func (m *MockPeer) String() string {
	//TODO: implement
	return ""
}

// Send mocks base method
func (m *MockPeer) Send(msgcode uint64, data interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", msgcode, data)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockPeerMockRecorder) Send(msgcode, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockPeer)(nil).Send), msgcode, data)
}
