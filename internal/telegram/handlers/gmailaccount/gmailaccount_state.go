package gmailaccount

import "sync"

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
