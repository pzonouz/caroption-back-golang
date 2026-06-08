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

func (s *Service) ListCheckBandsWithSortFilterPagination(
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

	WhereClause, args := utils.BuildWhere(
		"check_bands",
		filters,
		filterOperands,
		filterConditions,
		1,
	)
	OrderByClause := utils.BuildOrderBy("check_bands", sort, sortDirection)
	query := fmt.Sprintf(`
		SELECT
		    check_bands.id,
		    check_bands.name,
		    check_bands.serial,
		    check_bands.count,
		    check_bands.start_number,
		    check_bands.end_number,
		    check_bands.bank_id,
		    check_bands.used_numbers,
		    banks.name,
		    banks.branch,
		    check_bands.created_at,
		    check_bands.updated_at
		FROM
		    check_bands
		    LEFT JOIN banks ON banks.id = check_bands.bank_id %s %s %s`, WhereClause, OrderByClause, pagedBy)

	rows, err := s.db.Query(context.Background(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer rows.Close()

	var checkBands []CheckBand

	for rows.Next() {
		var checkBand CheckBand
		if err := rows.Scan(&checkBand.ID, &checkBand.Name, &checkBand.Serial, &checkBand.Count, &checkBand.StartNumber, &checkBand.EndNumber, &checkBand.BankId, &checkBand.UsedNumbers, &checkBand.BankName, &checkBand.BankBranch, &checkBand.CreatedAt, &checkBand.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		checkBands = append(checkBands, checkBand)
	}

	w.Header().Add("Content-Type", "application/json")

	newQuery := fmt.Sprintf(`
		SELECT
		    COUNT(*)
		FROM
		    check_bands %s`, WhereClause)
	row := s.db.QueryRow(context.Background(), newQuery, args...)

	var Count int32

	err = row.Scan(&Count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var checksWithTotalCount struct {
		Rows       []CheckBand `json:"rows"`
		TotalCount int32       `json:"totalCount"`
	}

	checksWithTotalCount.Rows = checkBands
	checksWithTotalCount.TotalCount = Count

	err = json.NewEncoder(w).Encode(checksWithTotalCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (s *Service) GetCheckBand(id string) (CheckBand, error) {
	var checkBand CheckBand

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return checkBand, err
	}

	query := `
		SELECT
		    check_bands.id,
		    check_bands.serial,
		    check_bands.count,
		    check_bands.start_number,
		    check_bands.end_number,
		    check_bands.used_numbers,
		    check_bands.bank_id,
		    check_bands.name,
		    banks.name,
		    banks.branch,
		    check_bands.created_at,
		    check_bands.updated_at
		FROM
		    check_bands
		    LEFT JOIN banks ON banks.id = check_bands.bank_id
		WHERE
		    check_bands.id = $1`
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(
		&checkBand.ID,
		&checkBand.Serial,
		&checkBand.Count,
		&checkBand.StartNumber,
		&checkBand.EndNumber,
		&checkBand.UsedNumbers,
		&checkBand.BankId,
		&checkBand.Name,
		&checkBand.BankName,
		&checkBand.BankBranch,
		&checkBand.CreatedAt,
		&checkBand.UpdatedAt,
	)
	if err != nil {
		return checkBand, err
	}

	return checkBand, nil
}

func (s *Service) CreateCheckBand(checkBand CheckBand) error {
	query := `
		SELECT
		    name
		FROM
		    banks
		WHERE
		    id = $1`
	row := s.db.QueryRow(context.Background(), query, checkBand.BankId)
	row.Scan(&checkBand.Name)

	query = `
		INSERT INTO check_bands (id, serial, name, count, start_number, end_number, used_numbers, bank_id)
		    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	validate := utils.NewValidate()

	err := validate.Struct(checkBand)
	if err != nil {
		return err
	}

	id := uuid.New()

	_, err = s.db.Exec(
		context.Background(),
		query,
		id,
		checkBand.Serial,
		fmt.Sprintf("%s-%s", checkBand.CreatedAt.String(), checkBand.BankId.String),
		checkBand.Count,
		checkBand.StartNumber,
		checkBand.EndNumber,
		[]string{},
		checkBand.BankId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditCheckBand(id string, checkBand CheckBand) error {
	query := `
		UPDATE
		    check_bands
		SET
		    serial = $1,
		    count = $2,
		    start_number = $3,
		    end_number = $4,
		    bank_id = $5
		WHERE
		    id = $6`

	validate := utils.NewValidate()

	err := validate.Struct(checkBand)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		context.Background(),
		query,
		checkBand.Serial,
		checkBand.Count,
		checkBand.StartNumber,
		checkBand.EndNumber,
		checkBand.BankId,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteCheckBand(id string) error {
	query := "DELETE FROM check_bands WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
