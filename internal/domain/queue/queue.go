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
	MatchedWithUserID int       //<-с кем нашелся поиск ?
	SearchRange       int       //<-рендж поиска по тем, у кого рейтнг+- наш
}

type Queue struct {
	Users              []UserInQueue
	Penalties          map[int]time.Time
	DefaultSearchRange int
	PenaltyDuration    time.Duration
	AcceptTimeOut      time.Duration
	reMu               sync.RWMutex
}

func NewQueue() *Queue {
	return &Queue{
		Users:              make([]UserInQueue, 0),
		Penalties:          make(map[int]time.Time),
		DefaultSearchRange: DefaultSearchRange,
		PenaltyDuration:    PenaltyDefaultMinutes,
		AcceptTimeOut:      AcceptTimeoutDefault,
	}
}

func (q *Queue) HasActivePenalty(userID int) bool {
	if q == nil {
		return false
	}
	if userID <= 0 {
		return false
	}
	q.reMu.RLock()
	defer q.reMu.RUnlock()
	penaltyUntil, ok := q.Penalties[userID]
	if !ok {
		return false
	}
	return time.Now().Before(penaltyUntil)
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
		return ErrBadUserID
	}
	q.reMu.Lock()
	defer q.reMu.Unlock()
	for i := range q.Users {
		if q.Users[i].UserID == user.UserID {
			return ErrUserAlreadyInQueue
		}
	}
	if penaltyUntil, ok := q.Penalties[user.UserID]; ok && time.Now().Before(penaltyUntil) {
		return ErrUserQueuePenalty
	}
	user.MatchedWithUserID = 0
	user.Status = QueueStatusSearching
	user.JoinedAt = time.Now()
	if user.SearchRange <= 0 {
		user.SearchRange = q.DefaultSearchRange
	}
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
		return ErrBadUserID
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
	foundCondidate := false
	bestDiff := 0
	for i := range q.Users {
		candidate := q.Users[i]
		if candidate.UserID == currentUser.UserID {
			continue
		}
		if candidate.Status != QueueStatusSearching {
			continue
		}
		diff := ratingDiff(currentUser.Rating, candidate.Rating)
		if diff > currentUser.SearchRange {
			continue
		}
		if !foundCondidate || diff < bestDiff {
			bestCandidate = candidate
			bestDiff = diff
			foundCondidate = true
		}
	}
	if !foundCondidate {
		return UserInQueue{}, false
	}
	return bestCandidate, true
}

// размер очереди
func (q *Queue) Size() int {
	if q == nil {
		return 0
	}
	q.reMu.RLock()
	defer q.reMu.RUnlock()
	return len(q.Users)
}

// проверям, может ли пользователь вообще зайти в очередь
func (q *Queue) CanJoinQueue(user *UserInQueue) error {
	if q == nil {
		return ErrNilQueue
	}
	if user == nil {
		return ErrNilUser
	}
	if user.UserID <= 0 {
		return ErrBadUserID
	}
	if q.ContainsUserInQueue(user.UserID) {
		return ErrUserAlreadyInQueue
	}
	if q.HasActivePenalty(user.UserID) {
		return ErrUserQueuePenalty
	}
	return nil
}

// обновляем статус пользователя в очереди (стоит в очереди, нашел матч, заблокирован и тд)
func (q *Queue) UpdateUserStatusInQueue(userID int,
	status string, mathedWithUser int) error {
	if q == nil {
		return ErrNilQueue
	}
	if userID <= 0 {
		return ErrBadUserID
	}
	if status == "" {
		return ErrBadStatus
	}
	q.reMu.Lock()
	defer q.reMu.Unlock()
	for i := range q.Users {
		if q.Users[i].UserID != userID {
			continue
		}
		switch status {
		case QueueStatusSearching:
			q.Users[i].Status = QueueStatusSearching
			q.Users[i].MatchedWithUserID = 0
			return nil
		case QueueStatusPendingMatch:
			if mathedWithUser <= 0 {
				return ErrBadUserID
			}
			q.Users[i].Status = QueueStatusPendingMatch
			q.Users[i].MatchedWithUserID = mathedWithUser
			return nil
		default:
			return ErrBadStatus
		}
	}
	return ErrUserNotFoundInQueue
}

// для штрафов, таймаутов и так далее
func (q *Queue) CleanupExpired() {
	if q == nil {
		return
	}
	q.reMu.Lock()
	defer q.reMu.Unlock()
	now := time.Now()
	for userID, penaltyUntil := range q.Penalties {
		if penaltyUntil.IsZero() || !now.Before(penaltyUntil) {
			delete(q.Penalties, userID)
		}
	}
}

func (q *Queue) ApplyPenalty(userID int, until time.Time) error {
	if q == nil {
		return ErrNilQueue
	}
	if userID <= 0 {
		return ErrBadUserID
	}
	if until.IsZero() || !time.Now().Before(until) {
		return ErrBadPenaltyTime
	}
	q.reMu.Lock()
	defer q.reMu.Unlock()
	for i := range q.Users {
		if q.Users[i].UserID == userID {
			q.Users = append(q.Users[:i], q.Users[i+1:]...)
			break
		}
	}
	q.Penalties[userID] = until
	return nil
}

func (q *Queue) GetPenaltyUntil(userID int) (time.Time, bool) {
	if q == nil {
		return time.Time{}, false
	}
	if userID <= 0 {
		return time.Time{}, false
	}
	q.reMu.RLock()
	defer q.reMu.RUnlock()
	penaltyUntil, ok := q.Penalties[userID]
	if !ok {
		return time.Time{}, false
	}
	if penaltyUntil.IsZero() || !time.Now().Before(penaltyUntil) {
		return time.Time{}, false
	}
	return penaltyUntil, true
}
