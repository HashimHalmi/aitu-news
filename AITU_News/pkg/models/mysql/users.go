package mysql

import (
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"hamedfrogh.net/aitunews/pkg/models"
	"strings"
)

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password, role string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	stmt := `INSERT INTO users (name, email, hashed_password, created, role)
	VALUES(?, ?, ?, UTC_TIMESTAMP(), ?)`

	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword), role)
	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return models.ErrDuplicateEmail
			}
		}
		return err
	}

	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte
	stmt := "SELECT id, hashed_password FROM users WHERE email = ? AND active = TRUE"
	row := m.DB.QueryRow(stmt, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, models.ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, models.ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil

}

func (m *UserModel) Get(id int) (*models.User, error) {
	return nil, nil
}

func (m *UserModel) GetRoleByID(id int) (string, error) {
	var role string
	stmt := "SELECT role FROM users WHERE id = ?"
	err := m.DB.QueryRow(stmt, id).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

func (m *UserModel) SetApprovalStatus(userID int, approved bool) error {
	stmt := "UPDATE users SET approved = ? WHERE id = ?"
	_, err := m.DB.Exec(stmt, approved, userID)
	if err != nil {
		return err
	}
	return nil
}

func (m *UserModel) IsApproved(userID int) (bool, error) {
	var approved bool
	stmt := "SELECT approved FROM users WHERE id = ?"
	err := m.DB.QueryRow(stmt, userID).Scan(&approved)
	if err != nil {
		return false, err
	}
	return approved, nil
}

func (m *UserModel) GetPendingTeachers() ([]*models.User, error) {
	// Fetch a list of teacher users pending approval from the database
	stmt := "SELECT id, name, email FROM users WHERE role = 'teacher' AND approved = FALSE"
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pendingTeachers []*models.User
	for rows.Next() {
		teacher := &models.User{}
		err := rows.Scan(&teacher.ID, &teacher.Name, &teacher.Email)
		if err != nil {
			return nil, err
		}
		pendingTeachers = append(pendingTeachers, teacher)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pendingTeachers, nil
}
