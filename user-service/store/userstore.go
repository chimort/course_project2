package store

import (
	"sync"

	"github.com/chimort/course_project2/models"
)

type UserStore struct {
	sync.Mutex
	users map[int64]*models.User
	nextID int64
}

func NewUserStore() *UserStore {
	return &UserStore{
		users:  make(map[int64]*models.User),
		nextID: 1,
	}
}

func (s *UserStore) AddUser(u *models.User) *models.User {
	s.Lock()
	defer s.Unlock()

	u.ID = s.nextID
	s.nextID++
	s.users[u.ID] = u
	return u
}

func (s *UserStore) GetUser(id int64) (*models.User, bool) {
	s.Lock()
	defer s.Unlock()

	u, ok := s.users[id]
	return u, ok
}