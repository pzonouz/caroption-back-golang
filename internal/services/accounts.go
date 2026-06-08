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

func (s *Service) ListAccountsWithSortFilterPagination(
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

	filterBy, args := utils.BuildWhere("accounts", filters, filterOperands, filterConditions, 1)
	sortBy := utils.BuildOrderBy("accounts", sort, sortDirection)
	query := fmt.Sprintf(`
		SELECT
		    accounts.id,
		    accounts.name,
		    accounts.parent_id,
		    p.name AS parent_name,
		    accounts.code,
		    accounts.created_at
		FROM
		    accounts
		    LEFT JOIN accounts p ON accounts.parent_id = p.id %s %s %s`, filterBy, sortBy, pagedBy)

	rows, err := s.db.Query(context.Background(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer rows.Close()

	var accounts []Account

	for rows.Next() {
		var account Account
		if err := rows.Scan(&account.ID, &account.Name, &account.ParentId, &account.ParentName, &account.Code, &account.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		accounts = append(accounts, account)
	}

	w.Header().Add("Content-Type", "application/json")

	newQuery := fmt.Sprintf(`
		SELECT
		    COUNT(*)
		FROM
		    accounts
		    LEFT JOIN accounts p ON accounts.parent_id = p.id %s`, filterBy)
	row := s.db.QueryRow(context.Background(), newQuery, args...)

	var Count int32

	err = row.Scan(&Count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var accountsWithTotalCount struct {
		Rows       []Account `json:"rows"`
		TotalCount int32     `json:"totalCount"`
	}

	accountsWithTotalCount.Rows = accounts
	accountsWithTotalCount.TotalCount = Count

	err = json.NewEncoder(w).Encode(accountsWithTotalCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (s *Service) ListAccounts() ([]Account, error) {
	query := `
		SELECT
		    a.id,
		    a.name,
		    a.parent_id,
		    p.name AS parent_name,
		    a.code,
		    a.locked,
		    a.created_at
		FROM
		    accounts AS a
		    LEFT JOIN accounts p ON a.parent_id = p.id`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []Account{}, err
	}
	defer rows.Close()

	var accounts []Account

	for rows.Next() {
		var account Account
		if err := rows.Scan(&account.ID, &account.Name, &account.ParentId, &account.ParentName, &account.Code, &account.Locked, &account.CreatedAt); err != nil {
			return []Account{}, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (s *Service) GetAccount(id string) (Account, error) {
	var account Account

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return account, err
	}

	query := `
		SELECT
		    a.id,
		    a.name,
		    a.parent_id,
		    p.name AS parent_name,
		    a.code,
		    a.locked,
		    a.created_at
		FROM
		    accounts AS a
		    LEFT JOIN accounts AS p ON p.id = a.parent_id
		WHERE
		    a.id = $1;
		
		`
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(
		&account.ID,
		&account.Name,
		&account.ParentId,
		&account.ParentName,
		&account.Code,
		&account.Locked,
		&account.CreatedAt,
	)
	if err != nil {
		return account, err
	}

	return account, nil
}

func (s *Service) ListParentAccounts() ([]Account, error) {
	query := `
		SELECT
		    p.id,
		    p.name,
		    p.parent_id,
		    p.code,
		    p.created_at,
		    COALESCE(json_agg(json_build_object('id', a.id, 'name', a.name, 'parentId', a.parent_id, 'code', a.code, 'createdAt', a.created_at)) FILTER (WHERE a.id IS NOT NULL), '[]') AS children
		FROM
		    accounts p
		    LEFT JOIN accounts a ON a.parent_id = p.id
		WHERE
		    p.parent_id IS NULL
		GROUP BY
		    p.id`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account

	for rows.Next() {
		var account Account
		if err := rows.Scan(
			&account.ID,
			&account.Name,
			&account.ParentId,
			&account.Code,
			&account.CreatedAt,
			&account.Children,
		); err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (s *Service) CreateAccount(account Account) error {
	query := "INSERT INTO accounts (id,name,parent_id,code) VALUES ($1,$2,$3,$4);"

	id := uuid.New()

	_, err := s.db.Exec(
		context.Background(),
		query,
		id,
		account.Name,
		account.ParentId,
		account.Code,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditAccount(id string, account Account) error {
	query := "UPDATE accounts SET name=$1,parent_id=$2,code=$3 WHERE id=$4;"
	validate := utils.NewValidate()

	err := validate.Struct(account)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		context.Background(),
		query,
		account.Name,
		account.ParentId,
		account.Code,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteAccount(id string) error {
	query := "DELETE FROM accounts WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
