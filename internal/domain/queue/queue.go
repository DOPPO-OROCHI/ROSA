package queue

import (
	"sync"
	"time"
)

type UserInQueue struct {
	UserID            int       //<-айди пользователя в очереди
	Rating            int       //<-его рейтинг
	JoinedAt          time.Time //<-время, которое игрок уже находится в очереди
	Status            string    //<-статус игрока в очереди (ищет, нашел оппонента, временно забанен)
	PenaltyUntil      time.Time //<-до какого времени чувак забанен ?
	MatchedWithUserID int       //<-с кем нашелся поиск ?
	SearchRange       int       //<-рендж поиска по тем, у кого рейтнг+- наш
}

type Queue struct {
	reMu               sync.RWMutex
	Users              []UserInQueue
	DefaultSearchRange int
	PenaltyDuration    int
	AcceptTimeOut      int
}

// добавляем пользователя в очередь
func (q *Queue) AddUserToQueue(user *UserInQueue) error {
	if q == nil {
		return ErrNilQueue
	}
	if user == nil {
		return ErrNilUser
	}
	if user.UserID <= 0 {
		return ErrInvalidUserID
	}
	q.reMu.Lock()
	defer q.reMu.Unlock()
	for i := range q.Users {
		usr := q.Users[i]
		if usr.UserID == user.UserID {
			return ErrUserAlreadyInQueue
		}
	}
	if !user.PenaltyUntil.IsZero() && time.Now().Before(user.PenaltyUntil) {
		return ErrUserQueuePenalty
	}
	user.MatchedWithUserID = 0
	user.Status = QueueStatusSearching
	user.JoinedAt = time.Now()
	q.Users = append(q.Users, *user)
	return nil
}

// убираем пользователя из очереди
func (q *Queue) RemoveUserFromQueue(user *UserInQueue) error {
	if q == nil {
		return ErrNilQueue
	}
	if user == nil {
		return ErrNilUser
	}
	if user.UserID <= 0 {
		return ErrInvalidUserID
	}
	q.reMu.Lock()
	defer q.reMu.Unlock()
	for i := range q.Users {
		if q.Users[i].UserID == user.UserID {
			q.Users = append(q.Users[:i], q.Users[i+1:]...)
			return nil
		}
	}
	return ErrUserNotFoundInQueue
}

// проверяем есть ли пользователь в очереди
func (q *Queue) ContainsUserInQueue(userID int) bool {
	if q == nil {
		return false
	}
	if userID <= 0 {
		return false
	}
	q.reMu.RLock()
	defer q.reMu.RUnlock()
	for i := range q.Users {
		usr := q.Users[i].UserID
		if usr == userID {
			return true
		}
	}
	return false
}

// получаем пользователя из очереди
func (q *Queue) GetUserInQueue(userID int) (UserInQueue, bool) {
	if q == nil {
		return UserInQueue{}, false
	}
	if userID <= 0 {
		return UserInQueue{}, false
	}
	q.reMu.RLock()
	defer q.reMu.RUnlock()
	for i := range q.Users {
		if q.Users[i].UserID == userID {
			return q.Users[i], true
		}
	}
	return UserInQueue{}, false
}

// ищем матч для пользователя
func (q *Queue) FindMatchForUser(userID int) (UserInQueue, bool) {
	if q == nil {
		return UserInQueue{}, false
	}
	if userID <= 0 {
		return UserInQueue{}, false
	}
	ratingDiff := func(a, b int) int {
		if a > b {
			return a - b
		}
		return b - a
	}
	q.reMu.RLock()
	defer q.reMu.RUnlock()
	var currentUser UserInQueue
	foundCurrent := false
	for i := range q.Users {
		if q.Users[i].UserID == userID {
			currentUser = q.Users[i]
			foundCurrent = true
			break
		}
	}
	if !foundCurrent {
		return UserInQueue{}, false
	}
	var bestCandidate UserInQueue
	foundCandidate := false
	bestDiff := 0
	now := time.Now()
	for i := range q.Users {
		candidate := q.Users[i]
		if candidate.UserID == currentUser.UserID {
			continue
		}
		if candidate.Status != QueueStatusSearching {
			continue
		}
		if !candidate.PenaltyUntil.IsZero() && now.Before(candidate.PenaltyUntil) {
			continue
		}
		diff := ratingDiff(currentUser.Rating, candidate.Rating)
		if diff > currentUser.SearchRange {
			continue
		}
		if !foundCandidate || diff < bestDiff {
			bestCandidate = candidate
			bestDiff = diff
			foundCandidate = true
		}
	}
	if !foundCandidate {
		return UserInQueue{}, false
	}
	return bestCandidate, true
}

// размер очереди
func (q *Queue) Size() {}

// проверям, может ли пользователь вообще зайти в очередь
func (q *Queue) CanJoinQueue() {}

// обновляем статус пользователя в очереди (стоит в очереди, нашел матч, заблокирован и тд)
func (q *Queue) UpdateUserStatusInQueue() {}

// для штрафов, таймаутов и так далее
func (q *Queue) CleanupExpired() {}
