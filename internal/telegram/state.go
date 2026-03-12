package telegram

import (
	"sync"

	"github.com/google/uuid"
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

// addPrompt conversation state

type addPromptStep int

const (
	stepAddPromptWaitFilter addPromptStep = iota
	stepAddPromptWaitPrompt
)

type addPromptState struct {
	step   addPromptStep
	editID *uuid.UUID // nil = new prompt; non-nil = editing existing
	filter string     // accumulated from step 1
}

var (
	addPromptMu     sync.Mutex
	addPromptStates = map[int64]*addPromptState{}
)

func getAddPromptState(userID int64) *addPromptState {
	addPromptMu.Lock()
	defer addPromptMu.Unlock()
	return addPromptStates[userID]
}

func setAddPromptState(userID int64, s *addPromptState) {
	addPromptMu.Lock()
	defer addPromptMu.Unlock()
	if s == nil {
		delete(addPromptStates, userID)
	} else {
		addPromptStates[userID] = s
	}
}

// setGmailAccount conversation state

type setGmailAccountState struct{}

var (
	setGmailAccountMu     sync.Mutex
	setGmailAccountStates = map[int64]*setGmailAccountState{}
)

func getSetGmailAccountState(userID int64) *setGmailAccountState {
	setGmailAccountMu.Lock()
	defer setGmailAccountMu.Unlock()
	return setGmailAccountStates[userID]
}

func setSetGmailAccountState(userID int64, s *setGmailAccountState) {
	setGmailAccountMu.Lock()
	defer setGmailAccountMu.Unlock()
	if s == nil {
		delete(setGmailAccountStates, userID)
	} else {
		setGmailAccountStates[userID] = s
	}
}
