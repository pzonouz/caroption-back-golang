package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) ListParametersWithSortFilterPagination(
	sort string,
	sortDirection string,
	filters []string,
	filterOperands []string,
	filterConditions []string,
	countInPage string,
	offset string,
	w http.ResponseWriter,
) {
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

	orderByClause := utils.BuildOrderBy("parameters", sort, sortDirection)

	whereClause, args := utils.BuildWhere(
		"parameters",
		filters,
		filterOperands,
		filterConditions,
		1,
	)

	query := fmt.Sprintf(`
		SELECT
		    parameters.id,
		    parameters.name,
		    parameters.description,
		    parameters.type,
		    parameters.parameter_group_id,
		    pgs.name,
			  parameters.selectables,
			  parameters.priority,
		    parameters.created_at
		FROM
		    parameters
		    LEFT JOIN public.parameter_groups AS pgs ON pgs.id = parameters.parameter_group_id
	   		%s %s %s`, whereClause, orderByClause, pagedBy)

	rows, err := s.db.Query(context.Background(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer rows.Close()

	var parameters []Parameter

	for rows.Next() {
		var parameter Parameter
		if err := rows.Scan(&parameter.ID, &parameter.Name, &parameter.Description, &parameter.Type, &parameter.ParameterGroupId, &parameter.ParameterGroup, &parameter.Selectables, &parameter.Priority, &parameter.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		parameters = append(parameters, parameter)
	}

	w.Header().Add("Content-Type", "application/json")

	newQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM
		    parameters 
		LEFT JOIN public.parameter_groups AS pgs 
ON pgs.id = parameters.parameter_group_id
		%s
		`, whereClause)
	row := s.db.QueryRow(context.Background(), newQuery, args...)

	var Count int32

	err = row.Scan(&Count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var parametersWithTotalCount struct {
		Rows       []Parameter `json:"rows"`
		TotalCount int32       `json:"totalCount"`
	}

	parametersWithTotalCount.Rows = parameters
	parametersWithTotalCount.TotalCount = Count

	err = json.NewEncoder(w).Encode(parametersWithTotalCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (s *Service) ListParametersByCategory(category_id string) ([]Parameter, error) {
	query := `
		SELECT
		    p.id,
		    p.name,
		    p.description,
		    p.type,
		    p.parameter_group_id,
		    p.selectables,
		    p.priority,
		    p.created_at
		FROM
		    parameters AS p
		    JOIN parameter_groups AS pg ON pg.id = p.parameter_group_id
		    JOIN categories AS c ON c.id = pg.category_id
		WHERE
		    c.id IN (
		        SELECT
		            cc.parent_id
		        FROM
		            categories AS cc
		        WHERE
		            cc.id = $1)
		ORDER BY
		    p.priority`

	rows, err := s.db.Query(context.Background(), query, category_id)
	if err != nil {
		return []Parameter{}, err
	}
	defer rows.Close()

	var parameters []Parameter

	for rows.Next() {
		var parameter Parameter
		if err := rows.Scan(&parameter.ID, &parameter.Name, &parameter.Description, &parameter.Type, &parameter.ParameterGroupId, &parameter.Selectables, &parameter.Priority, &parameter.CreatedAt); err != nil {
			return []Parameter{}, err
		}

		parameters = append(parameters, parameter)
	}

	return parameters, nil
}

func (s *Service) GetParameter(id string) (Parameter, error) {
	var parameter Parameter

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return parameter, err
	}

	query := `
	SELECT
    p.id,
    p.name,
    p.description,
    p.type,
    p.parameter_group_id,
    p.selectables,
    p.priority,
    p.created_at
FROM
    parameters p
LEFT JOIN
    parameter_groups pg ON p.parameter_group_id = pg.id
WHERE
    p.id = $1
		`
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(
		&parameter.ID,
		&parameter.Name,
		&parameter.Description,
		&parameter.Type,
		&parameter.ParameterGroupId,
		&parameter.Selectables,
		&parameter.Priority,
		&parameter.CreatedAt,
	)
	if err != nil {
		return parameter, err
	}

	return parameter, nil
}

func (s *Service) CreateParameter(parameter Parameter) error {
	query := `
		INSERT INTO parameters (id, name, description, type, parameter_group_id, selectables, priority)
		    VALUES ($1, $2, $3, $4, $5, $6, $7)`
	validate := utils.NewValidate()

	err := validate.Struct(parameter)
	if err != nil {
		return err
	}

	id := uuid.New()

	_, err = s.db.Exec(
		context.Background(),
		query,
		id,
		parameter.Name,
		parameter.Description,
		parameter.Type,
		parameter.ParameterGroupId,
		parameter.Selectables,
		parameter.Priority,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditParameter(id string, parameter Parameter) error {
	query := `
		UPDATE
		    parameters
		SET
		    name = $1,
		    description = $2,
		    type = $3,
		    parameter_group_id = $4,
		    selectables = $5,
		    priority = $6
		WHERE
		    id = $7`
	validate := utils.NewValidate()

	err := validate.Struct(parameter)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		context.Background(),
		query,
		parameter.Name,
		parameter.Description,
		parameter.Type,
		parameter.ParameterGroupId,
		parameter.Selectables,
		parameter.Priority,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteParameter(id string) error {
	query := "DELETE FROM parameters WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
