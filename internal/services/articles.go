package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) ListArticles() ([]Article, error) {
	query := `
		SELECT
		    a.id,
		    a.name,
		    a.description,
		    a.category_id,
		    a.slug,
		    a.keywords,
		    a.created_at,
		    a.updated_at,
		    a.show_in_products,
		    a.image_id,
		    i.image_url
		FROM
		    articles a
		    LEFT JOIN images i ON a.image_id = i.id`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []Article{}, err
	}
	defer rows.Close()

	var articles []Article

	for rows.Next() {
		var article Article

		if err := rows.Scan(&article.ID, &article.Name, &article.Description, &article.CategoryID, &article.Slug, &article.Keywords, &article.CreatedAt, &article.UpdatedAt, &article.ShowInProducts, &article.ImageID, &article.ImageUrl); err != nil {
			return []Article{}, err
		}

		articles = append(articles, article)
	}

	return articles, nil
}

func (s *Service) GetArticle(id string) (Article, error) {
	var article Article

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return article, err
	}

	query := `
		SELECT
		    a.id,
		    a.name,
		    a.description,
		    a.category_id,
		    a.slug,
		    a.keywords,
		    a.created_at,
		    a.updated_at,
		    a.show_in_products,
		    a.image_id,
		    i.image_url,
		FROM
		    articles a
		    LEFT JOIN images i ON a.image_id = i.id
		WHERE
		    a.id = $1;
		
		`
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	var articleParameterValuesJSON []byte

	err = row.Scan(
		&article.ID,
		&article.Name,
		&article.Description,
		&article.CategoryID,
		&article.Slug,
		&article.Keywords,
		&article.CreatedAt,
		&article.UpdatedAt,
		&article.ShowInProducts,
		&article.ImageID,
		&article.ImageUrl,
		&articleParameterValuesJSON,
	)
	if err != nil {
		return Article{}, err
	}

	return article, nil
}

func (s *Service) GetArticleBySlug(slug string) (Article, error) {
	var article Article

	query := `
		SELECT
		    a.id,
		    a.name,
		    a.description,
		    a.category_id,
		    a.slug,
		    a.keywords,
		    a.created_at,
		    a.updated_at,
		    a.show_in_products,
		    a.image_id,
		    i.image_url
		FROM
		    articles a
		    LEFT JOIN images i ON a.image_id = i.id
		WHERE
		    a.slug = $1;
		
		`
	row := s.db.QueryRow(context.Background(), query, slug)

	err := row.Scan(
		&article.ID,
		&article.Name,
		&article.Description,
		&article.CategoryID,
		&article.Slug,
		&article.Keywords,
		&article.CreatedAt,
		&article.UpdatedAt,
		&article.ShowInProducts,
		&article.ImageID,
		&article.ImageUrl,
	)
	if err != nil {
		return article, err
	}

	return article, nil
}

func (s *Service) CreateArticle(article Article) error {
	query := `
		INSERT INTO articles (id, name, description, slug, keywords, category_id, image_id, show_in_products)
		    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	validate := utils.NewValidate()

	err := validate.Struct(article)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	id := uuid.New()

	_, err = tx.Exec(
		context.Background(),
		query,
		id,
		article.Name,
		article.Description,
		article.Slug,
		article.Keywords,
		article.CategoryID,
		article.ImageID,
		article.ShowInProducts,
	)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func (s *Service) EditArticle(id string, article Article) error {
	query := "UPDATE articles SET name=$1,description=$2,image_id=$3,slug=$4,keywords=$5,category_id=$6,show_in_products=$7 WHERE id=$8;"
	validate := utils.NewValidate()

	err := validate.Struct(article)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = s.db.Exec(
		context.Background(),
		query,
		article.Name,
		article.Description,
		article.ImageID,
		article.Slug,
		article.Keywords,
		article.CategoryID,
		article.ShowInProducts,
		id,
	)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func (s *Service) DeleteArticle(id string) error {
	query := "DELETE FROM articles WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ArticlesInCategory(category_id string) ([]Article, error) {
	query := `
		SELECT
		    a.id,
		    a.name,
		    a.description,
		    a.category_id,
		    a.slug,
		    a.keywords,
		    a.created_at,
		    a.updated_at,
		    a.show_in_products,
		    a.image_id,
		    i.image_url
		FROM
		    articles a
		    LEFT JOIN images i ON a.image_id = i.id
		    LEFT JOIN categories c ON c.id = a.category_id
		WHERE
		    c.parent_id = $1
		    OR c.id = $1;
		
		`

	rows, err := s.db.Query(context.Background(), query, category_id)
	if err != nil {
		return []Article{}, err
	}
	defer rows.Close()

	var articles []Article

	for rows.Next() {
		var article Article
		if err := rows.Scan(&article.ID, &article.Name, &article.Description, &article.CategoryID, &article.Slug, &article.Keywords, &article.CreatedAt, &article.UpdatedAt, &article.ShowInProducts, &article.ImageID, &article.ImageUrl); err != nil {
			return []Article{}, err
		}

		articles = append(articles, article)
	}

	return articles, nil
}

func (s *Service) RecentlyArticles() ([]Article, error) {
	delayedTime := time.Now().AddDate(0, 0, -30)
	query := `
		SELECT
		    p.id,
		    p.name,
		    p.description,
		    p.slug,
		    p.keywords,
		    p.created_at,
		    p.updated_at,
		    p.image_id,
		    i.image_url
		FROM
		    articles p
		    LEFT JOIN images i ON p.image_id = i.id
		WHERE
		    p.created_at > $1
		ORDER BY
		    p.created_at ASC`

	rows, err := s.db.Query(context.Background(), query, delayedTime)
	if err != nil {
		return []Article{}, err
	}
	defer rows.Close()

	var articles []Article

	for rows.Next() {
		var article Article

		if err := rows.Scan(&article.ID, &article.Name, &article.Description, &article.Slug, &article.Keywords, &article.CreatedAt, &article.UpdatedAt, &article.ImageID, &article.ImageUrl); err != nil {
			return []Article{}, err
		}

		articles = append(articles, article)
	}

	return articles, nil
}
