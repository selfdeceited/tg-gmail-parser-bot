package addprompt

import (
	"sync"

	"github.com/google/uuid"
)

type addPromptStep int

const (
	stepWaitFilter addPromptStep = iota
	stepWaitPrompt
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
