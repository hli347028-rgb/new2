package data

import (
	"context"
	"sync"
	"time"

	"backend/internal/biz"
)

type challengeRepo struct {
	mu    sync.RWMutex
	items map[string]*biz.Challenge
}

// NewChallengeRepo creates an in-memory challenge repository.
func NewChallengeRepo() biz.ChallengeRepo {
	return &challengeRepo{items: make(map[string]*biz.Challenge)}
}

func (r *challengeRepo) Save(_ context.Context, challenge *biz.Challenge) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[challenge.Address] = challenge
	return nil
}

func (r *challengeRepo) Get(_ context.Context, address string) (*biz.Challenge, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	challenge, ok := r.items[address]
	if !ok {
		return nil, nil
	}
	if time.Now().After(challenge.ExpireAt) {
		return nil, nil
	}
	return challenge, nil
}

func (r *challengeRepo) Delete(_ context.Context, address string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, address)
	return nil
}
