package queue

import (
	"sort"
	"sync"
	"time"
)

/*
1 -ебануть структуру очереди
2 -описать поля (что мы ждем от модуля ?)
3 -
*/

type UserInQueue struct {
	UserID   int
	Rating   int
	JoinedAt time.Time
}

type Queue struct {
	Users []UserInQueue
	reMu  sync.RWMutex
}

func (q *Queue) Filter() {
	//очередь должна уметь :
	//сортировать пользователей по рейтингу
	//уметь добавлять пользователя в слайс юзеров в очереди
	sort.Slice(q.Users, func(i, j int) bool {
		return q.Users[i].rating > q.Users[j].rating
	})
}

func (i *Queue) AddUsersToQueue(user *UserInQueue) {
	i.Users = append(i.Users, *user)
}

func (q *Queue) MarryUsers() {}

func (q *Queue) deleteUserFromQueue(userID int) error {
	return nil
}

//защита от ДДОС
//должна быть задержка перед началом участия в поиске очереди, дать Лехе ответ почему

//подсказка -месячные

//написать тесты для этог пакета
