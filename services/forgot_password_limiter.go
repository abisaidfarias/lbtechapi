package services

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"go.mongodb.org/mongo-driver/mongo"
)

type forgotPasswordRateLimiter struct {
	mu      sync.Mutex
	entries map[string]time.Time
}

var forgotPasswordLimiter = &forgotPasswordRateLimiter{entries: map[string]time.Time{}}

func (l *forgotPasswordRateLimiter) allow(key string, minInterval time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	last, ok := l.entries[key]
	if ok && now.Sub(last) < minInterval {
		return false
	}
	l.entries[key] = now
	return true
}

func (s *authService) sendForgotPasswordIfExists(email string) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return
	}
	key := "email:" + email
	if !forgotPasswordLimiter.allow(key, time.Minute) {
		log.Printf("forgot-password rate limited for %s", email)
		return
	}
	user, err := s.userRepository.GetByEmail(email)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Printf("forgot-password lookup failed: %v", err)
		}
		return
	}
	name := strings.TrimSpace(user.Name + " " + user.LastName)
	if err := s.passwordTokenService.IssueAndEmail(user.ID, user.Email, name, models.PasswordTokenPurposeReset); err != nil {
		log.Printf("forgot-password email failed for %s: %v", user.Email, err)
	}
}
