package repositories

import (
	"database/sql"
	"switchiot/internal/db"
	"switchiot/internal/domain/entities"
	"switchiot/internal/domain/repositories"
)

// SQLUserRepository implements UserRepository using SQL database
type SQLUserRepository struct {
	db *sql.DB
}

// NewSQLUserRepository creates a new SQL user repository
func NewSQLUserRepository(database *sql.DB) repositories.UserRepository {
	return &SQLUserRepository{db: database}
}

// Create creates a new user
func (r *SQLUserRepository) Create(username, passwordHash, role string) (int64, error) {
	return db.CreateUser(r.db, username, passwordHash, role)
}

// GetByUsername returns a user by username along with password hash
func (r *SQLUserRepository) GetByUsername(username string) (*entities.User, string, error) {
	dbUser, passwordHash, found, err := db.GetUserByUsername(r.db, username)
	if err != nil {
		return nil, "", err
	}
	if !found {
		return nil, "", sql.ErrNoRows
	}

	user := &entities.User{
		ID:        dbUser.ID,
		Username:  dbUser.Username,
		Role:      dbUser.Role,
		CreatedAt: dbUser.CreatedAt,
	}

	return user, passwordHash, nil
}

// GetAll returns all users
func (r *SQLUserRepository) GetAll() ([]entities.User, error) {
	dbUsers, err := db.ListUsers(r.db)
	if err != nil {
		return nil, err
	}

	users := make([]entities.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = entities.User{
			ID:        dbUser.ID,
			Username:  dbUser.Username,
			Role:      dbUser.Role,
			CreatedAt: dbUser.CreatedAt,
		}
	}

	return users, nil
}

// Delete deletes a user by ID
func (r *SQLUserRepository) Delete(id int64) error {
	return db.DeleteUser(r.db, id)
}

// Count returns the total number of users
func (r *SQLUserRepository) Count() (int, error) {
	return db.CountUsers(r.db)
}