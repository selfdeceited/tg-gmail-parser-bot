package service

import (
	"context"
	"html"
	"net/mail"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/claude"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/commands"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/entities"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/queries"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/gmail"
)

const pollInterval = 120 * time.Second
const pollTimeout = 15 * time.Second

// SendFunc delivers a formatted message to a Telegram chat.
type SendFunc func(chatID int64, msg string)

// WatchService manages per-user Gmail polling goroutines.
type WatchService interface {
	// Start begins watching for a user. Idempotent — silently replaces an existing watcher.
	Start(ctx context.Context, userID int64, chatID int64, send SendFunc) error
	// Stop cancels an active watcher. Returns true if one was running.
	Stop(userID int64) bool
	// IsWatching reports whether a watcher goroutine is active for the user.
	IsWatching(userID int64) bool
	// RestoreAll resumes watchers for all users flagged is_watching=true in DB.
	// Called once on server startup.
	RestoreAll(ctx context.Context, send SendFunc) error
	// Wait blocks until all poller goroutines have exited.
	// Call after the root context is cancelled for a clean shutdown.
	Wait()
}

type watchService struct {
	db        *gorm.DB
	claude    *claude.Client
	ioTimeout time.Duration

	mu      sync.Mutex
	cancels map[int64]context.CancelFunc
	wg      sync.WaitGroup
}

// NewWatchService returns a WatchService backed by the given DB and Claude client.
func NewWatchService(db *gorm.DB, claudeClient *claude.Client, ioTimeout time.Duration) WatchService {
	return &watchService{
		db:        db,
		claude:    claudeClient,
		ioTimeout: ioTimeout,
		cancels:   make(map[int64]context.CancelFunc),
	}
}

func (s *watchService) Start(ctx context.Context, userID int64, chatID int64, send SendFunc) error {
	// Persist watch state first so a restart can resume.
	if err := commands.SetWatching(s.db, userID, chatID, true); err != nil {
		return err
	}

	s.mu.Lock()
	// Cancel any existing watcher for this user before starting a new one.
	if cancel, ok := s.cancels[userID]; ok {
		cancel()
	}
	watchCtx, cancel := context.WithCancel(ctx)
	s.cancels[userID] = cancel
	s.mu.Unlock()

	s.wg.Add(1)
	go s.runLoop(watchCtx, userID, chatID, send)
	logrus.WithFields(logrus.Fields{"user_id": userID, "chat_id": chatID}).Info("watch: started")
	return nil
}

func (s *watchService) Stop(userID int64) bool {
	s.mu.Lock()
	cancel, ok := s.cancels[userID]
	if ok {
		cancel()
		delete(s.cancels, userID)
	}
	s.mu.Unlock()

	if ok {
		if err := commands.SetWatching(s.db, userID, 0, false); err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("watch: failed to persist stop — watcher may resume on restart")
		}
		logrus.WithField("user_id", userID).Info("watch: stopped")
	}
	return ok
}

func (s *watchService) IsWatching(userID int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.cancels[userID]
	return ok
}

func (s *watchService) Wait() {
	s.wg.Wait()
}

func (s *watchService) RestoreAll(ctx context.Context, send SendFunc) error {
	users, err := queries.GetWatchingUsers(s.db)
	if err != nil {
		return err
	}
	// consider in-db pagination for larger datasets
	for _, u := range users {
		if err := s.Start(ctx, u.ID, u.WatchChatID, send); err != nil {
			logrus.WithError(err).WithField("user_id", u.ID).Error("watch: failed to restore watcher on startup")
		}
	}
	logrus.WithField("count", len(users)).Info("watch: restored watchers from DB")
	return nil
}

func (s *watchService) runLoop(ctx context.Context, userID int64, chatID int64, send SendFunc) {
	defer s.wg.Done()
	log := logrus.WithFields(logrus.Fields{"user_id": userID, "chat_id": chatID})

	since := s.loadLastChecked(userID, log)
	log.WithField("since", since).Info("watch: poll loop starting")

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("watch: poll loop stopped")
			return
		case <-ticker.C:
			now := time.Now().UTC()
			log.WithField("since", since).Info("watch: polling Gmail")
			pollCtx, pollCancel := context.WithTimeout(ctx, pollTimeout)
			err := s.poll(pollCtx, userID, chatID, since, send, log)
			pollCancel()
			if err != nil {
				log.WithError(err).Error("watch: poll failed, last_checked_at not advanced")
			} else {
				since = now
				if err := commands.UpdateLastChecked(s.db, userID, since); err != nil {
					log.WithError(err).Error("watch: failed to update last_checked_at")
				}
			}
		}
	}
}

func (s *watchService) loadLastChecked(userID int64, log *logrus.Entry) time.Time {
	user, err := queries.GetUser(s.db, userID)
	if err != nil {
		log.WithError(err).Warn("watch: failed to load last_checked_at, defaulting to now")
		return time.Now().UTC()
	}
	if user.LastCheckedAt == nil {
		return time.Now().UTC()
	}
	return *user.LastCheckedAt
}

func (s *watchService) poll(ctx context.Context, userID int64, chatID int64, since time.Time, send SendFunc, log *logrus.Entry) error {
	ioCtx, cancel := context.WithTimeout(ctx, s.ioTimeout)
	defer cancel()

	db := s.db.WithContext(ioCtx)

	creds, err := queries.GetCredentials(db, userID)
	if err != nil {
		log.WithError(err).Error("watch: failed to load credentials")
		return err
	}

	gmailService, err := gmail.NewGmailService(ioCtx, creds.ClientID, creds.ClientSecret, creds.RefreshToken)
	if err != nil {
		log.WithError(err).Error("watch: failed to create gmail service")
		return err
	}

	user, err := queries.GetUser(db, userID)
	if err != nil {
		log.WithError(err).Error("watch: failed to load user, aborting poll")
		return err
	}
	accountIndex := user.GmailAccountIndex

	emails, err := gmail.FetchNewMessages(ioCtx, gmailService, since, accountIndex)
	if err != nil {
		log.WithError(err).Error("watch: failed to fetch messages")
		return err
	}

	prompts, err := queries.GetActivePrompts(db, userID)
	if err != nil {
		log.WithError(err).Error("watch: failed to load prompts")
		return err
	}
	if len(prompts) == 0 {
		log.Info("watch: no prompts configured, skipping")
		return nil
	}

	for _, email := range emails {
		if ctx.Err() != nil {
			log.WithError(ctx.Err()).Warn("watch: poll context expired, stopping email processing")
			break
		}
		s.processEmail(ctx, userID, chatID, email, prompts, send, log)
	}
	return nil
}

func (s *watchService) processEmail(ctx context.Context, userID int64, chatID int64, email gmail.EmailMessage, prompts []entities.Prompt, send SendFunc, log *logrus.Entry) {
	log = log.WithFields(logrus.Fields{
		"message_id": email.ID,
		"user_id":    userID,
		"from":       email.From,
		"subject":    email.Subject,
	})
	log.Info("watch: processing email")

	filtered := filterPrompts(email, prompts)
	log.WithField("prompt_count", len(filtered)).Info("watch: prompts selected for email")

	for _, p := range filtered {
		log = log.WithFields(logrus.Fields{
			"prompt_id": p.ID,
			"email_URL": email.URL,
		})
		result, err := s.claude.Summarize(ctx, p.Prompt, email)
		if err != nil {
			log.WithError(err).Error("watch: claude summarization failed, skipping prompt")
			continue
		}

		if strings.EqualFold(result.Result, "matched") {
			log.Info("watch: email matched the prompt, sending to chat")
			send(chatID, formatSummary(result, email.URL, p.ID.String()[:6]))
			return
		}

		log.Info("watch: did not match the prompt, skipping")
	}

	log.Info("watch: no prompt matched, email ignored")
}

// filterPrompts returns the prompts to run against this email.
// If any prompt filter matches the sender, only that prompt is returned (first-match intent).
// Otherwise only prompts with no filter are returned.
func filterPrompts(email gmail.EmailMessage, prompts []entities.Prompt) []entities.Prompt {
	fromAddr := extractEmailAddress(email.From)
	var selected []entities.Prompt
	for _, p := range prompts {
		if strings.EqualFold(p.Filter, fromAddr) || strings.EqualFold(p.Filter, email.From) {
			return []entities.Prompt{p} // spec: "first parser with the matching sender address"
		} else if p.Filter == "" {
			selected = append(selected, p)
		}
	}
	return selected
}

// extractEmailAddress parses the bare email address from an RFC 5322 From header
// value such as "Pointer <suraj@pointer.io>". Falls back to the raw string if
// parsing fails.
func extractEmailAddress(from string) string {
	addr, err := mail.ParseAddress(from)
	if err != nil {
		return from
	}
	return addr.Address
}

func formatSummary(r *claude.SummarizeResult, url, promptShortID string) string {
	return "📧 <b>" + html.EscapeString(r.Title) + "</b>\n\n" +
		html.EscapeString(r.ContentString()) +
		"\n\n<a href=\"" + url + "\">Open in Gmail</a>" +
		"\\|  matched by <code>" + promptShortID + "</code>"
}
