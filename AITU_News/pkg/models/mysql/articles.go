package mysql

import (
	"context"
	"database/sql"
	"errors"

	"hamedfrogh.net/aitunews/pkg/models"
)

type ArticleModel struct {
	DB *sql.DB
}

func (m *ArticleModel) Insert(title, content, expires, category string) (int, error) {

	stmt := `INSERT INTO articles (title, content, created, expires, category)
    VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY), ?)`

	result, err := m.DB.Exec(stmt, title, content, expires, category)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, nil
	}
	return int(id), nil
}

func (m *ArticleModel) Get(id int) (*models.Article, error) {

	stmt := `SELECT id, title, content, created, expires, category FROM articles
    WHERE expires > UTC_TIMESTAMP() AND id = ?`

	row := m.DB.QueryRow(stmt, id)

	s := &models.Article{}

	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires, &s.Category)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}

	return s, nil
}

func (m *ArticleModel) Latest(ctx context.Context) ([]*models.Article, error) {

	stmt := `SELECT id, title, content, created, expires, category FROM articles
    WHERE expires > UTC_TIMESTAMP() ORDER BY created DESC LIMIT 10`

	rows, err := m.DB.QueryContext(ctx, stmt)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	articles := []*models.Article{}
	for rows.Next() {
		s := &models.Article{}
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires, &s.Category)
		if err != nil {
			return nil, err
		}
		articles = append(articles, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

func (m *ArticleModel) GetByCategory(category string) ([]*models.Article, error) {
	stmt := `SELECT id, title, content, created, expires, category FROM articles
    WHERE expires > UTC_TIMESTAMP() AND category = ? ORDER BY created DESC`

	rows, err := m.DB.Query(stmt, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := []*models.Article{}
	for rows.Next() {
		s := &models.Article{}
		err := rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires, &s.Category)
		if err != nil {
			return nil, err
		}
		articles = append(articles, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

func (m *ArticleModel) GetCategories() ([]string, error) {
	stmt := `SELECT DISTINCT category FROM articles`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		err := rows.Scan(&category)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}
