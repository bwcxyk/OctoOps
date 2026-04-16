package rbac

import (
	"errors"
	"sync"
)

// PasswordResetStore is a storage abstraction for reset-code and reset-rate data.
// Future Redis implementation can satisfy this interface and be injected via SetPasswordResetStore.
type PasswordResetStore interface {
	GetCode(email string) (resetCodeEntry, bool, error)
	SetCode(email string, entry resetCodeEntry) error
	DeleteCode(email string) error

	GetRate(key string) (resetRateEntry, bool, error)
	SetRate(key string, entry resetRateEntry) error
}

type memoryPasswordResetStore struct {
	mu    sync.RWMutex
	codes map[string]resetCodeEntry
	rates map[string]resetRateEntry
}

func NewMemoryPasswordResetStore() PasswordResetStore {
	return &memoryPasswordResetStore{
		codes: make(map[string]resetCodeEntry),
		rates: make(map[string]resetRateEntry),
	}
}

var passwordResetStore PasswordResetStore = NewMemoryPasswordResetStore()

func SetPasswordResetStore(store PasswordResetStore) error {
	if store == nil {
		return errors.New("password reset store cannot be nil")
	}
	passwordResetStore = store
	return nil
}

func (s *memoryPasswordResetStore) GetCode(email string) (resetCodeEntry, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.codes[email]
	return entry, ok, nil
}

func (s *memoryPasswordResetStore) SetCode(email string, entry resetCodeEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[email] = entry
	return nil
}

func (s *memoryPasswordResetStore) DeleteCode(email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.codes, email)
	return nil
}

func (s *memoryPasswordResetStore) GetRate(key string) (resetRateEntry, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.rates[key]
	return entry, ok, nil
}

func (s *memoryPasswordResetStore) SetRate(key string, entry resetRateEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rates[key] = entry
	return nil
}
