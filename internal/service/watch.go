package service

import (
	"context"
	"html"
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

const pollInterval = 20 * time.Second

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
}

type watchService struct {
	db     *gorm.DB
	claude *claude.Client

	mu      sync.Mutex
	cancels map[int64]context.CancelFunc
}

// NewWatchService returns a WatchService backed by the given DB and Claude client.
func NewWatchService(db *gorm.DB, claudeClient *claude.Client) WatchService {
	return &watchService{
		db:      db,
		claude:  claudeClient,
		cancels: make(map[int64]context.CancelFunc),
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
		_ = commands.SetWatching(s.db, userID, 0, false)
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
	log := logrus.WithFields(logrus.Fields{"user_id": userID, "chat_id": chatID})

	since := s.loadLastChecked(userID)
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
			if err := s.poll(ctx, userID, chatID, since, send, log); err != nil {
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

func (s *watchService) loadLastChecked(userID int64) time.Time {
	user, err := queries.GetUser(s.db, userID)
	if err != nil || user.LastCheckedAt == nil {
		return time.Now().UTC()
	}
	return *user.LastCheckedAt
}

func (s *watchService) poll(ctx context.Context, userID int64, chatID int64, since time.Time, send SendFunc, log *logrus.Entry) error {
	creds, err := queries.GetCredentials(s.db, userID)
	if err != nil {
		log.WithError(err).Error("watch: failed to load credentials")
		return err
	}

	gmailSvc, err := gmail.NewGmailService(ctx, creds.ClientID, creds.ClientSecret, creds.RefreshToken)
	if err != nil {
		log.WithError(err).Error("watch: failed to create gmail service")
		return err
	}

	emails, err := gmail.FetchNewMessages(ctx, gmailSvc, since)
	if err != nil {
		log.WithError(err).Error("watch: failed to fetch messages")
		return err
	}

	prompts, err := queries.GetActivePrompts(s.db, userID)
	if err != nil {
		log.WithError(err).Error("watch: failed to load prompts")
		return err
	}
	if len(prompts) == 0 {
		log.Info("watch: no prompts configured, skipping")
		return nil
	}

	for _, email := range emails {
		s.processEmail(ctx, userID, chatID, email, prompts, send, log)
	}
	return nil
}

func (s *watchService) processEmail(ctx context.Context, userID int64, chatID int64, email gmail.EmailMessage, prompts []entities.Prompt, send SendFunc, log *logrus.Entry) {
	log = log.WithFields(logrus.Fields{
		"message_id": email.ID,
		"from":       email.From,
		"subject":    email.Subject,
	})
	log.Info("watch: processing email")

	toTry := selectPrompts(email, prompts)
	log.WithField("prompt_count", len(toTry)).Info("watch: prompts selected for email")

	for _, p := range toTry {
		result, err := s.claude.Summarize(ctx, p.Prompt, email)
		if err != nil {
			log.WithError(err).WithField("prompt_id", p.ID).Error("watch: claude summarization failed, skipping prompt")
			continue
		}

		if strings.EqualFold(result.Result, "matched") {
			log.WithField("prompt_id", p.ID).Info("watch: email matched, sending to chat")
			send(chatID, formatSummary(result))
			return
		}

		log.WithField("prompt_id", p.ID).Info("watch: email not matched, trying next prompt")
	}

	log.Info("watch: no prompt matched, email ignored")
}

// selectPrompts returns the prompts to run against this email.
// If any prompt filter matches the sender, only those are returned (first-match intent).
// Otherwise all prompts are returned in order.
func selectPrompts(email gmail.EmailMessage, prompts []entities.Prompt) []entities.Prompt {
	var matched []entities.Prompt
	for _, p := range prompts {
		if p.Filter != "" && strings.EqualFold(p.Filter, email.From) {
			matched = append(matched, p)
		}
	}
	if len(matched) > 0 {
		return matched[:1] // spec: "first parser with the matching sender address"
	}
	return prompts
}

func formatSummary(r *claude.SummarizeResult) string {
	return "📧 <b>" + html.EscapeString(r.Title) + "</b>\n\n" + html.EscapeString(r.ContentString())
}
