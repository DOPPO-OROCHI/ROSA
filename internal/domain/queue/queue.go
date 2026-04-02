package queue

import (
	"strconv"
	"sync"
	"time"
)

/*Данный файл полностью посвящен описанию такой сущности как очередь. В моей игре очередь работает внутри
оперативной памяти. Это сделано потому, что мой учитель так сказал. Почему не брать инфу из БД ? Вопросы
тут не тебе задавать... Так и вот. Чтобы реализовать такую сущность как очередь, я прибегнул к структурам,
методы которой и будут отвечать за операционку внутри очередей.*/

// структура пользователя внутри очереди, грубо говоря слепок объекта, который находится в массиве очереди
type UserInQueue struct {
	UserID            int       //<-айди пользователя в очереди
	Rating            int       //<-его рейтинг
	JoinedAt          time.Time //<-время, которое игрок уже находится в очереди
	Status            string    //<-статус игрока в очереди (ищет, нашел оппонента, временно забанен)
	MatchedWithUserID int       //<-с кем нашелся поиск ?
	SearchRange       int       //<-рендж поиска по тем, у кого рейтнг+- наш
}

// а это очередь, где находятся игроки, а так же поля, который нужны для...
type Queue struct {
	Users              []UserInQueue
	Penalties          map[int]time.Time //<-подсчета штрафников (если игрок не принял игру-штраф 3 минуты)
	DefaultSearchRange int               //<-дефолтный рендж по поиску (типа по рейтингу матчим игроков)
	PenaltyDuration    time.Duration     //<-длительность штрафа отдельно взятого игрока
	AcceptTimeout      time.Duration     //<-таймаут на принятие матча
	PendingMatches     map[string]PendingMatch
	PendingByUser      map[int]string
	reMu               sync.RWMutex //<-мьютекс
}

type PendingMatch struct {
	ID              string
	UserID1         int
	UserID2         int
	AcceptedByUser1 bool
	AcceptedByUser2 bool
	CreatedAt       time.Time
	ExpiresAt       time.Time
}

func NewQueue() *Queue {
	return &Queue{
		Users:              make([]UserInQueue, 0),
		Penalties:          make(map[int]time.Time),
		DefaultSearchRange: DefaultSearchRange,
		PenaltyDuration:    PenaltyDefaultMinutes,
		AcceptTimeout:      AcceptTimeoutDefault,
		PendingMatches:     make(map[string]PendingMatch),
		PendingByUser:      make(map[int]string),
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
			if penaltyUntil, ok := q.Penalties[userID]; ok && time.Now().Before(penaltyUntil) {
				return UserInQueue{}, false
			}
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
		if diff > candidate.SearchRange {
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

// для штрафов, таймаутов и так далее
func (q *Queue) CleanupExpired() {
	if q == nil {
		return
	}
	q.reMu.Lock()
	defer q.reMu.Unlock()
	now := time.Now()
	toRemove := make(map[int]bool)
	for userID, penaltyUntil := range q.Penalties {
		if penaltyUntil.IsZero() || !now.Before(penaltyUntil) {
			delete(q.Penalties, userID)
		}
	}
	for pendingID, pending := range q.PendingMatches {
		if now.Before(pending.ExpiresAt) {
			continue
		}
		idx1 := -1
		idx2 := -1
		for i := range q.Users {
			if q.Users[i].UserID == pending.UserID1 {
				idx1 = i
			}
			if q.Users[i].UserID == pending.UserID2 {
				idx2 = i
			}
		}
		if pending.AcceptedByUser1 == false {
			q.Penalties[pending.UserID1] = now.Add(q.PenaltyDuration)
			toRemove[pending.UserID1] = true
		}
		if pending.AcceptedByUser1 == true && idx1 != -1 {
			q.Users[idx1].Status = QueueStatusSearching
			q.Users[idx1].MatchedWithUserID = 0
		}
		if pending.AcceptedByUser2 == false {
			q.Penalties[pending.UserID2] = now.Add(q.PenaltyDuration)
			toRemove[pending.UserID2] = true
		}
		if pending.AcceptedByUser2 == true && idx2 != -1 {
			q.Users[idx2].Status = QueueStatusSearching
			q.Users[idx2].MatchedWithUserID = 0
		}
		delete(q.PendingMatches, pendingID)
		delete(q.PendingByUser, pending.UserID1)
		delete(q.PendingByUser, pending.UserID2)
	}
	if len(toRemove) > 0 {
		filtered := q.Users[:0]
		for _, user := range q.Users {
			if toRemove[user.UserID] {
				continue
			}
			filtered = append(filtered, user)
		}
		q.Users = filtered
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

func (q *Queue) ReserveMatchForUser(userID int) (UserInQueue, UserInQueue, bool, error) {
	if q == nil {
		return UserInQueue{}, UserInQueue{}, false, ErrNilQueue
	}
	if userID <= 0 {
		return UserInQueue{}, UserInQueue{}, false, ErrBadUserID
	}
	q.reMu.Lock()
	defer q.reMu.Unlock()
	currentIdx := -1
	bestCandidateIdx := -1
	bestDiff := 0
	ratingDiff := func(a, b int) int {
		if a > b {
			return a - b
		}
		return b - a
	}
	for i := range q.Users {
		if q.Users[i].UserID == userID {
			currentIdx = i
			break
		}
	}
	if currentIdx == -1 {
		return UserInQueue{}, UserInQueue{}, false, ErrUserNotFoundInQueue
	}
	if q.Users[currentIdx].Status != QueueStatusSearching {
		return UserInQueue{}, UserInQueue{}, false, ErrBadStatus
	}
	if _, ok := q.PendingByUser[userID]; ok {
		return UserInQueue{}, UserInQueue{}, false, ErrBadStatus
	}
	if penaltyUntil, ok := q.Penalties[userID]; ok && time.Now().Before(penaltyUntil) {
		return UserInQueue{}, UserInQueue{}, false, ErrUserQueuePenalty
	}
	for i := range q.Users {
		if i == currentIdx {
			continue
		}
		candidate := q.Users[i]
		if candidate.Status != QueueStatusSearching {
			continue
		}
		diff := ratingDiff(q.Users[currentIdx].Rating, candidate.Rating)
		if diff > q.Users[currentIdx].SearchRange {
			continue
		}
		if diff > candidate.SearchRange {
			continue
		}
		if bestCandidateIdx == -1 || diff < bestDiff {
			bestCandidateIdx = i
			bestDiff = diff
		}
	}
	if bestCandidateIdx == -1 {
		return UserInQueue{}, UserInQueue{}, false, nil
	}
	if _, ok := q.PendingByUser[q.Users[bestCandidateIdx].UserID]; ok {
		return UserInQueue{}, UserInQueue{}, false, ErrBadStatus
	}
	now := time.Now()
	pendingID := strconv.FormatInt(now.UnixNano(), 10)
	pending := PendingMatch{
		ID:              pendingID,
		UserID1:         q.Users[currentIdx].UserID,
		UserID2:         q.Users[bestCandidateIdx].UserID,
		AcceptedByUser1: false,
		AcceptedByUser2: false,
		CreatedAt:       now,
		ExpiresAt:       now.Add(q.AcceptTimeout),
	}
	q.PendingMatches[pendingID] = pending
	q.PendingByUser[q.Users[currentIdx].UserID] = pendingID
	q.PendingByUser[q.Users[bestCandidateIdx].UserID] = pendingID
	q.Users[currentIdx].Status = QueueStatusPendingMatch
	q.Users[currentIdx].MatchedWithUserID = q.Users[bestCandidateIdx].UserID
	q.Users[bestCandidateIdx].Status = QueueStatusPendingMatch
	q.Users[bestCandidateIdx].MatchedWithUserID = q.Users[currentIdx].UserID
	return q.Users[currentIdx], q.Users[bestCandidateIdx], true, nil
}

func (q *Queue) ReturnUserToSearch(userID int) error {
	if q == nil {
		return ErrNilQueue
	}
	if userID <= 0 {
		return ErrBadUserID
	}
	q.reMu.Lock()
	defer q.reMu.Unlock()
	for i := range q.Users {
		if q.Users[i].UserID == userID {
			q.Users[i].Status = QueueStatusSearching
			q.Users[i].MatchedWithUserID = 0
			return nil
		}
	}
	return ErrUserNotFoundInQueue
}

func (q *Queue) ResetPendingPair(userID1, userID2 int) error {
	if q == nil {
		return ErrNilQueue
	}
	if userID1 <= 0 || userID2 <= 0 || userID1 == userID2 {
		return ErrBadUserID
	}
	q.reMu.Lock()
	defer q.reMu.Unlock()
	idx1 := -1
	idx2 := -1
	for i := range q.Users {
		if q.Users[i].UserID == userID1 {
			idx1 = i
		}
		if q.Users[i].UserID == userID2 {
			idx2 = i
		}
	}
	if idx1 == -1 || idx2 == -1 {
		return ErrUserNotFoundInQueue
	}
	if q.Users[idx1].Status != QueueStatusPendingMatch || q.Users[idx2].Status != QueueStatusPendingMatch {
		return ErrBadStatus
	}
	if q.Users[idx1].MatchedWithUserID != userID2 || q.Users[idx2].MatchedWithUserID != userID1 {
		return ErrBadStatus
	}
	q.Users[idx1].Status = QueueStatusSearching
	q.Users[idx1].MatchedWithUserID = 0
	q.Users[idx2].Status = QueueStatusSearching
	q.Users[idx2].MatchedWithUserID = 0
	return nil
}

func (q *Queue) GetSearchDuration(userID int) (time.Duration, bool) {
	if q == nil {
		return 0, false
	}
	if userID <= 0 {
		return 0, false
	}
	q.reMu.RLock()
	defer q.reMu.RUnlock()
	for i := range q.Users {
		if q.Users[i].UserID == userID {
			if q.Users[i].JoinedAt.IsZero() {
				return 0, false
			}
			return time.Since(q.Users[i].JoinedAt), true
		}
	}
	return 0, false
}

func (q *Queue) GetUserMatchmakingState(userID int) string {
	if q == nil {
		return MatchMakingStateIdle
	}
	if userID <= 0 {
		return MatchMakingStateIdle
	}
	if q.HasActivePenalty(userID) {
		return MatchMakingStatePenalty
	}
	user, ok := q.GetUserInQueue(userID)
	if !ok {
		return MatchMakingStateIdle
	}
	switch user.Status {
	case QueueStatusSearching:
		return MatchMakingStateSearching
	case QueueStatusPendingMatch:
		return MatchMakingStatePendingMatch
	default:
		return MatchMakingStateIdle
	}
}

func (q *Queue) AcceptPendingMatch(userID int) (PendingMatch, bool, error) {
	if q == nil {
		return PendingMatch{}, false, ErrNilQueue
	}
	if userID <= 0 {
		return PendingMatch{}, false, ErrBadUserID
	}
	q.reMu.Lock()
	defer q.reMu.Unlock()
	pendingID, ok := q.PendingByUser[userID]
	if !ok {
		return PendingMatch{}, false, ErrBadStatus
	}
	pending, ok := q.PendingMatches[pendingID]
	if !ok {
		return PendingMatch{}, false, ErrBadStatus
	}
	if !time.Now().Before(pending.ExpiresAt) {
		return PendingMatch{}, false, ErrBadStatus
	}
	switch userID {
	case pending.UserID1:
		pending.AcceptedByUser1 = true
	case pending.UserID2:
		pending.AcceptedByUser2 = true
	default:
		return PendingMatch{}, false, ErrBadUserID
	}
	bothAccepted := pending.AcceptedByUser1 && pending.AcceptedByUser2
	q.PendingMatches[pendingID] = pending
	return pending, bothAccepted, nil
}

func (q *Queue) DeclinePendingMatch(userID int) error {
	if q == nil {
		return ErrNilQueue
	}
	if userID <= 0 {
		return ErrBadUserID
	}
	q.reMu.Lock()
	defer q.reMu.Unlock()
	pendingID, ok := q.PendingByUser[userID]
	if !ok {
		return ErrBadStatus
	}
	pending, ok := q.PendingMatches[pendingID]
	if !ok {
		return ErrBadStatus
	}
	if !time.Now().Before(pending.ExpiresAt) {
		return ErrBadStatus
	}
	declinerID := userID
	otherUserID := 0
	if pending.UserID1 == userID {
		otherUserID = pending.UserID2
	} else if pending.UserID2 == userID {
		otherUserID = pending.UserID1
	} else {
		return ErrBadUserID
	}
	declinerIdx := -1
	otherIdx := -1
	for i := range q.Users {
		if q.Users[i].UserID == declinerID {
			declinerIdx = i
		}
		if q.Users[i].UserID == otherUserID {
			otherIdx = i
		}
	}
	if declinerIdx == -1 || otherIdx == -1 {
		return ErrUserNotFoundInQueue
	}
	q.Users[otherIdx].Status = QueueStatusSearching
	q.Users[otherIdx].MatchedWithUserID = 0
	q.Users = append(q.Users[:declinerIdx], q.Users[declinerIdx+1:]...)
	q.Penalties[declinerID] = time.Now().Add(q.PenaltyDuration)
	delete(q.PendingMatches, pendingID)
	delete(q.PendingByUser, declinerID)
	delete(q.PendingByUser, otherUserID)
	return nil
}

func (q *Queue) FinalizeAcceptedMatch(userID1, userID2 int) error {
	if q == nil {
		return ErrNilQueue
	}
	if userID1 <= 0 || userID2 <= 0 {
		return ErrBadUserID
	}
	if userID1 == userID2 {
		return ErrBadUserID
	}
	q.reMu.Lock()
	defer q.reMu.Unlock()
	if pendingID, ok := q.PendingByUser[userID1]; ok {
		delete(q.PendingMatches, pendingID)
	} else if pendingID, ok := q.PendingByUser[userID2]; ok {
		delete(q.PendingMatches, pendingID)
	}
	delete(q.PendingByUser, userID1)
	delete(q.PendingByUser, userID2)
	filtered := q.Users[:0]
	for _, user := range q.Users {
		if user.UserID == userID1 || user.UserID == userID2 {
			continue
		}
		filtered = append(filtered, user)
	}
	q.Users = filtered
	return nil
}

func (q *Queue) GetPendingMatchInfo(userID int) (PendingMatch, error) {
	if q == nil {
		return PendingMatch{}, ErrNilQueue
	}
	if userID <= 0 {
		return PendingMatch{}, ErrBadUserID
	}
	q.reMu.RLock()
	defer q.reMu.RUnlock()
	pendingID, ok := q.PendingByUser[userID]
	if !ok {
		return PendingMatch{}, ErrBadStatus
	}
	pending, ok := q.PendingMatches[pendingID]
	if !ok {
		return PendingMatch{}, ErrBadStatus
	}
	return pending, nil
}

func (q *Queue) GetPendingAcceptanceState(userID int) (acceptedByMe bool,
	acceptedByOpponent bool, expiresAt time.Time, opponentUserID int, ok bool) {
	if q == nil {
		return false, false, time.Time{}, 0, false
	}
	if userID <= 0 {
		return false, false, time.Time{}, 0, false
	}
	pending, err := q.GetPendingMatchInfo(userID)
	if err != nil {
		return false, false, time.Time{}, 0, false
	}
	if userID == pending.UserID1 {
		return pending.AcceptedByUser1, pending.AcceptedByUser2, pending.ExpiresAt, pending.UserID2, true
	}
	if userID == pending.UserID2 {
		return pending.AcceptedByUser2, pending.AcceptedByUser1, pending.ExpiresAt, pending.UserID1, true
	}
	return false, false, time.Time{}, 0, false
}
