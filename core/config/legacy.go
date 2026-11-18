package config

import (
	"fmt"
	"strings"
	"time"

	"pantheon/core/domain"
)

// Всё в этом файле обслуживает пульт 0.1.x, который адресует фильтры индексом
// в срезе и сам привязывает первых написавших. Уходит вместе со старым пультом.

func (c *Config) GetSearches() []domain.Search {
	c.mu.RLock()
	defer c.mu.RUnlock()

	res := make([]domain.Search, len(c.Searches))
	copy(res, c.Searches)
	return res
}

func (c *Config) GetActiveSearches() []domain.Search {
	c.mu.RLock()
	defer c.mu.RUnlock()

	res := make([]domain.Search, 0, len(c.Searches))
	for _, s := range c.Searches {
		if !s.Disabled {
			res = append(res, s)
		}
	}
	return res
}

func (c *Config) AddSearch(s domain.Search) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if s.CreatedAt == 0 {
		s.CreatedAt = time.Now().Unix()
	}
	c.Searches = append(c.Searches, s)
	return c.saveLocked()
}

func (c *Config) RemoveSearch(index int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if index < 0 || index >= len(c.Searches) {
		return fmt.Errorf("неверный индекс фильтра: %d", index)
	}
	c.Searches = append(c.Searches[:index], c.Searches[index+1:]...)
	return c.saveLocked()
}

func (c *Config) ToggleSearch(index int) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if index < 0 || index >= len(c.Searches) {
		return false, fmt.Errorf("неверный индекс фильтра: %d", index)
	}
	c.Searches[index].Disabled = !c.Searches[index].Disabled
	state := !c.Searches[index].Disabled
	return state, c.saveLocked()
}

// SetSearchOwner закрепляет фильтр за конкретным чатом (владельцем алертов).
// Пустой chatID делает фильтр общим: алерты уходят всем получателям.
func (c *Config) SetSearchOwner(index int, chatID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if index < 0 || index >= len(c.Searches) {
		return fmt.Errorf("неверный индекс фильтра: %d", index)
	}
	c.Searches[index].OwnerChatID = chatID
	return c.saveLocked()
}

// IsAuthorized — прежнее имя IsTrusted.
func (c *Config) IsAuthorized(chatID string) bool { return c.IsTrusted(chatID) }

// AutoAuthorizeChatID — авто-привязка первых пользователей бота.
func (c *Config) AutoAuthorizeChatID(chatID string, maxValid int) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	chatID = strings.TrimSpace(chatID)
	valid := 0
	for _, id := range c.TelegramChatIDs {
		if !ValidChatID(id) {
			continue
		}
		if id == chatID {
			return false, nil
		}
		valid++
	}
	if maxValid > 0 && valid >= maxValid {
		return false, fmt.Errorf("лимит доверенных пользователей (%d)", maxValid)
	}

	c.TelegramChatIDs = append(c.TelegramChatIDs, chatID)
	return true, c.saveLocked()
}
