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

func (s *Service) ListChecksWithSortFilterPagination(
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

	orderByClause := utils.BuildOrderBy("checks", sort, sortDirection)

	whereClause, args := utils.BuildWhere(
		"checks",
		filters,
		filterOperands,
		filterConditions,
		1,
	)
	query := fmt.Sprintf(`
		SELECT
		    checks.id,
		    checks.name,
		    checks.type,
		    checks.check_number,
		    checks.bank_name,
		    checks.sayyad,
		    checks.amount,
		    checks.description,
		    checks.due_date,
		    checks.person_id,
		    CONCAT(persons.name, ' ', persons.first_name) AS person_name,
		    checks.status,
		    checks.check_band_id,
		    checks.created_at,
		    checks.updated_at
		FROM
		    checks
		    LEFT JOIN persons ON checks.person_id = persons.id %s %s %s`, whereClause, orderByClause, pagedBy)

	rows, err := s.db.Query(context.Background(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer rows.Close()

	var checks []Check

	for rows.Next() {
		var check Check
		if err := rows.Scan(&check.ID, &check.Name, &check.Type, &check.CheckNumber, &check.BankName, &check.Sayyad, &check.Amount, &check.Description, &check.DueDate, &check.PersonId, &check.PersonName, &check.Status, &check.CheckBandId, &check.CreatedAt, &check.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		checks = append(checks, check)
	}

	w.Header().Add("Content-Type", "application/json")

	newQuery := fmt.Sprintf(`
		SELECT
		    COUNT(*)
		FROM
		    checks %s`, whereClause)
	row := s.db.QueryRow(context.Background(), newQuery, args...)

	var Count int32

	err = row.Scan(&Count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var checksWithTotalCount struct {
		Rows       []Check `json:"rows"`
		TotalCount int32   `json:"totalCount"`
	}

	checksWithTotalCount.Rows = checks
	checksWithTotalCount.TotalCount = Count

	err = json.NewEncoder(w).Encode(checksWithTotalCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (s *Service) GetCheck(id string) (Check, error) {
	var check Check

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return check, err
	}
	query := `
		SELECT
		    checks.id,
		    checks.name,
		    checks.type,
		    checks.check_number,
		    checks.bank_name,
		    checks.sayyad,
		    checks.amount,
		    checks.description,
		    checks.due_date,
		    checks.person_id,
		    checks.status,
		    checks.check_band_id,
		    checks.created_at,
		    checks.updated_at
		FROM
		    checks
		WHERE
		    id = $1`

	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(
		&check.ID,
		&check.Name,
		&check.Type,
		&check.CheckNumber,
		&check.BankName,
		&check.Sayyad,
		&check.Amount,
		&check.Description,
		&check.DueDate,
		&check.PersonId,
		&check.Status,
		&check.CheckBandId,
		&check.CreatedAt,
		&check.UpdatedAt,
	)
	if err != nil {
		return check, err
	}

	return check, nil
}

func (s *Service) CreateCheck(check Check) error {
	query := `
		INSERT INTO checks (id, name, type, check_number, bank_name, sayyad, amount, description, due_date, person_id, status, check_band_id)
		    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	validate := utils.NewValidate()

	err := validate.Struct(check)
	if err != nil {
		return err
	}

	id := uuid.New()

	_, err = s.db.Exec(
		context.Background(),
		query,
		id,
		fmt.Sprintf("%s-%s-%s", check.BankName.String, check.Amount.String, check.DueDate.String()),
		check.Type,
		check.CheckNumber,
		check.BankName,
		check.Sayyad,
		check.Amount,
		check.Description,
		check.DueDate,
		check.PersonId,
		check.Status,
		check.CheckBandId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditCheck(id string, check Check) error {
	query := `
		UPDATE
		    checks
		SET
		    name = $1,
		    type = $2,
		    check_number = $3,
		    bank_name = $4,
		    sayyad = $5,
		    amount = $6,
		    description = $7,
		    due_date = $8,
		    person_id = $9,
		    status = $10,
		    check_band_id = $11
		WHERE
		    id = $12`
	validate := utils.NewValidate()

	err := validate.Struct(check)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		context.Background(),
		query,
		fmt.Sprintf("%s-%s-%s", check.BankName.String, check.Amount.String, check.DueDate.String()),
		check.Type,
		check.CheckNumber,
		check.BankName,
		check.Sayyad,
		check.Amount,
		check.Description,
		check.DueDate,
		check.PersonId,
		check.Status,
		check.CheckBandId,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteCheck(id string) error {
	query := "DELETE FROM checks WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
