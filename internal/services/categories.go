package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) ListParentCategoriesForAccounting() ([]Category, error) {
	query := `
		SELECT
		    p.id,
		    p.name,
		    p.parent_id,
		    p.description,
		    p.priority,
		    i.image_url,
		    p.image_id,
		    p.slug,
		    p.show,
		    p.created_at,
		    p.updated_at,
		    COALESCE(json_agg(json_build_object('id', c.id, 'name', c.name, 'parentId', c.parent_id, 'slug', c.slug, 'createdAt', c.created_at)) FILTER (WHERE c.id IS NOT NULL), '[]') AS children
		FROM
		    categories p
		    LEFT JOIN categories c ON c.parent_id = p.id
		    LEFT JOIN images i ON p.image_id = i.id
		WHERE
		    p.parent_id IS NULL
		GROUP BY
		    p.id,
		    p.name,
		    p.parent_id,
		    p.description,
		    p.priority,
		    p.image_id,
		    p.slug,
		    p.show,
		    p.generator,
		    p.created_at,
		    p.updated_at,
		    i.image_url
		ORDER BY
		    p.priority`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category

	for rows.Next() {
		var category Category
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.ParentID,
			&category.Description,
			&category.Priority,
			&category.ImageUrl,
			&category.ImageID,
			&category.Slug,
			&category.Show,
			&category.CreatedAt,
			&category.UpdatedAt,
			&category.Children,
		); err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *Service) ListCategories() ([]Category, error) {
	query := `
		SELECT
		    c.id,
		    c.name,
		    c.parent_id,
		    p.name AS parent_name,
		    c.description,
		    c.priority,
		    c.image_id,
		    i.image_url,
		    c.slug,
		    c.show,
		    c.created_at,
		    c.updated_at
		FROM
		    categories AS c
		    LEFT JOIN categories p ON c.parent_id = p.id
		    LEFT JOIN images i ON c.image_id = i.id
		GROUP BY
		    c.id,
		    i.image_url,
		    p.name;
		
		`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []Category{}, err
	}
	defer rows.Close()

	var categories []Category

	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.ParentID, &category.ParentName, &category.Description, &category.Priority, &category.ImageID, &category.ImageUrl, &category.Slug, &category.Show, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return []Category{}, err
		}

		category.Children = []Child{}

		categories = append(categories, category)
	}

	return categories, nil
}

func (s *Service) ListParentCategories() ([]Category, error) {
	query := `
		SELECT
		    p.id,
		    p.name,
		    p.parent_id,
		    p.description,
		    p.priority,
		    p.slug,
		    p.created_at,
		    p.updated_at
		FROM
		    categories p
		WHERE
		    p.show IS TRUE
		    AND p.parent_id IS NULL
		ORDER BY
		    COALESCE(CAST(NULLIF (p.priority, '') AS bigint), 0)`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category

	for rows.Next() {
		var category Category
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.ParentID,
			&category.Description,
			&category.Priority,
			&category.Slug,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *Service) ListCategoriesForBlocks() ([]Category, error) {
	query := `
		SELECT
		    categories.id,
		    categories.name,
		    categories.parent_id,
		    categories.description,
		    categories.priority,
		    i.image_url,
		    categories.image_id,
		    categories.slug,
		    categories.show,
		    categories.created_at,
		    categories.updated_at
		FROM
		    categories
		    LEFT JOIN images i ON categories.image_id = i.id
		WHERE
		    categories.show IS TRUE
		ORDER BY
		    COALESCE(CAST(NULLIF (categories.priority, '') AS bigint), 0)`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category

	for rows.Next() {
		var category Category
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.ParentID,
			&category.Description,
			&category.Priority,
			&category.ImageUrl,
			&category.ImageID,
			&category.Slug,
			&category.Show,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *Service) ListCategoriesWithSortFilterPagination(
	sort string,
	sortDirection string,
	filters []string,
	filterOperands []string,
	filterConditions []string,
	countInPage string,
	offset string,
	w http.ResponseWriter,
) {
	orderBy := ""

	if sort != "" {
		switch sort {
		case "priority":
			orderBy = `ORDER BY COALESCE(CAST(NULLIF(categories.priority, '') AS bigint), 0)`
		case "show":
			orderBy = fmt.Sprintf(
				`ORDER BY categories.show %s`,
				sortDirection,
			)
		default:
			orderBy = fmt.Sprintf(
				`ORDER BY categories.%s COLLATE "fa-IR-x-icu" %s`,
				sort,
				sortDirection,
			)
		}
	}

	var filterBy strings.Builder
	if len(filters) > 0 {
		filterBy.WriteString("WHERE ")
	}

	for index, filter := range filters {
		filterOperand := filterOperands[index]
		filterCondition := filterConditions[index]

		if filterOperand == "contains" {
			filterOperand = "ILIKE"

			filterCondition = "%" + filterCondition + "%"
		}
		if filterOperand == "notcontains" {
			filterOperand = "NOT ILIKE"

			filterCondition = "%" + filterCondition + "%"
		}

		if len(filters) != 0 {
			switch filter {
			case "show":
				if filterCondition == "true" {
					filterBy.WriteString("categories.show IS TRUE")
				}
				if filterCondition == "false" {
					filterBy.WriteString("categories.show IS FALSE")
				}
			default:
				filterBy.WriteString(fmt.Sprintf(
					`%s %s '%s'`,
					"categories."+filter,
					filterOperand,
					filterCondition,
				))
			}
		}

		if len(filters)-1 > index {
			filterBy.WriteString(" AND ")
		}
	}

	pagedBy := ""
	offsetNum := 0

	if countInPage != "" {
		limit, err := strconv.Atoi(countInPage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		if offset != "" {
			offsetNum, err = strconv.Atoi(offset)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return
			}
		}

		pagedBy = fmt.Sprintf(`LIMIT %d OFFSET %d`, limit, offsetNum)
	}

	query := fmt.Sprintf(`
		SELECT
		    categories.id,
		    categories.name,
		    categories.parent_id,
		    p.name AS parent_name,
		    categories.description,
		    categories.priority,
		    categories.image_id,
		    images.image_url,
		    categories.slug,
		    categories.show,
		    categories.created_at,
		    categories.updated_at
		FROM
		    categories
		    LEFT JOIN categories p ON categories.parent_id = p.id
		    LEFT JOIN images ON categories.image_id = images.id %s %s %s`, filterBy.String(), orderBy, pagedBy)

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer rows.Close()

	var categories []Category

	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.ParentID, &category.ParentName, &category.Description, &category.Priority, &category.ImageID, &category.ImageUrl, &category.Slug, &category.Show, &category.CreatedAt, &category.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		category.Children = []Child{}

		categories = append(categories, category)
	}

	w.Header().Add("Content-Type", "application/json")

	newQuery := fmt.Sprintf(`
		SELECT
		    COUNT(*)
		FROM
		    categories
		    LEFT JOIN categories p ON categories.parent_id = p.id
		    LEFT JOIN images ON categories.image_id = images.id %s`, filterBy.String())
	row := s.db.QueryRow(context.Background(), newQuery)

	var Count int32

	err = row.Scan(&Count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var categoriesWithTotalCount struct {
		Rows       []Category `json:"rows"`
		TotalCount int32      `json:"totalCount"`
	}

	categoriesWithTotalCount.Rows = categories
	categoriesWithTotalCount.TotalCount = Count

	err = json.NewEncoder(w).Encode(categoriesWithTotalCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (s *Service) GetCategory(id string) (Category, error) {
	var category Category

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return category, err
	}

	query := `
		SELECT
		    c.id,
		    c.name,
		    c.parent_id,
		    p.name AS parent_name,
		    c.description,
		    c.priority,
		    c.image_id,
		    i.image_url,
		    c.slug,
		    c.show,
		    c.created_at,
		    c.updated_at
		FROM
		    categories AS c
		    LEFT JOIN categories p ON c.parent_id = p.id
		    LEFT JOIN images i ON c.image_id = i.id
		WHERE
		    c.id = $1;
		
		`
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(
		&category.ID,
		&category.Name,
		&category.ParentID,
		&category.ParentName,
		&category.Description,
		&category.Priority,
		&category.ImageID,
		&category.ImageUrl,
		&category.Slug,
		&category.Show,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return category, err
	}

	return category, nil
}

func (s *Service) GetCategoryBySlug(slug string) (Category, error) {
	var category Category

	query := `
		SELECT
		    id,
		    name,
		    parent_id,
		    description,
		    priority,
		    image_id,
		    slug,
		    SHOW,
		    created_at,
		    updated_at
		FROM
		    categories
		WHERE
		    slug = $1`
	row := s.db.QueryRow(context.Background(), query, slug)

	err := row.Scan(
		&category.ID,
		&category.Name,
		&category.ParentID,
		&category.Description,
		&category.Priority,
		&category.ImageID,
		&category.Slug,
		&category.Show,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return category, err
	}

	return category, nil
}

func (s *Service) CreateCategory(category Category) error {
	query := "INSERT INTO categories (id,name,parent_id,description,priority,image_id,slug,show) VALUES ($1,$2,$3,$4,$5,$6,$7,$8);"
	validate := utils.NewValidate()

	err := validate.Struct(category)
	if err != nil {
		return err
	}

	id := uuid.New()

	_, err = s.db.Exec(
		context.Background(),
		query,
		id,
		category.Name,
		category.ParentID,
		category.Description,
		category.Priority,
		category.ImageID,
		category.Slug,
		category.Show,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditCategory(id string, category Category) error {
	query := "UPDATE categories SET name=$1,description=$2,parent_id=$3,image_id=$4,priority=$5,slug=$6,show=$7 WHERE id=$8"
	validate := utils.NewValidate()

	err := validate.Struct(category)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		context.Background(),
		query,
		category.Name,
		category.Description,
		category.ParentID,
		category.ImageID,
		category.Priority,
		category.Slug,
		category.Show,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteCategory(id string) error {
	query := "DELETE FROM categories WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
