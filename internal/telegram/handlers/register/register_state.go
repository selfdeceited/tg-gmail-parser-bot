package register

import (
	"sync"

	"golang.org/x/oauth2"
)

type registerStep int

const (
	stepWaitCredentials registerStep = iota
	stepWaitAuthCode
)

type registerState struct {
	step        registerStep
	oauthConfig *oauth2.Config
}

var (
	registerMu     sync.Mutex
	registerStates = map[int64]*registerState{}
)

func getState(userID int64) *registerState {
	registerMu.Lock()
	defer registerMu.Unlock()
	return registerStates[userID]
}

func setState(userID int64, s *registerState) {
	registerMu.Lock()
	defer registerMu.Unlock()
	if s == nil {
		delete(registerStates, userID)
	} else {
		registerStates[userID] = s
	}
}
