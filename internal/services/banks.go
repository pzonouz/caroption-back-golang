package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) ListBanksWithSortFilterPagination(
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

	WhereClause, args := utils.BuildWhere("banks", filters, filterOperands, filterConditions, 1)
	OrderByClause := utils.BuildOrderBy("banks", sort, sortDirection)

	query := fmt.Sprintf(`
		SELECT
		    banks.id,
		    banks.name,
		    banks.branch,
		    banks.number,
		    banks.shaba,
		    banks.type,
		    banks.first_balance,
		    banks.created_at,
		    banks.updated_at
		FROM
		    banks %s %s %s`, WhereClause, OrderByClause, pagedBy)

	rows, err := s.db.Query(context.Background(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer rows.Close()

	var banks []Bank

	for rows.Next() {
		var bank Bank
		if err := rows.Scan(&bank.ID, &bank.Name, &bank.Branch, &bank.Number, &bank.Shaba, &bank.Type, &bank.FirstBalance, &bank.CreatedAt, &bank.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		banks = append(banks, bank)
	}

	w.Header().Add("Content-Type", "application/json")

	newQuery := fmt.Sprintf(`
		SELECT
		    COUNT(*)
		FROM
		    banks %s`, WhereClause)
	row := s.db.QueryRow(context.Background(), newQuery, args...)

	var Count int32

	err = row.Scan(&Count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var banksWithTotalCount struct {
		Rows       []Bank `json:"rows"`
		TotalCount int32  `json:"totalCount"`
	}

	banksWithTotalCount.Rows = banks
	banksWithTotalCount.TotalCount = Count

	err = json.NewEncoder(w).Encode(banksWithTotalCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (s *Service) GetBank(id string) (Bank, error) {
	var bank Bank

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return bank, err
	}

	query := `
		SELECT
		    id,
		    name,
		    branch,
		    number,
		    shaba,
		    type,
		    first_balance,
		    created_at
		FROM
		    banks
		WHERE
		    id = $1`
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(
		&bank.ID,
		&bank.Name,
		&bank.Branch,
		&bank.Number,
		&bank.Shaba,
		&bank.Type,
		&bank.FirstBalance,
		&bank.CreatedAt,
	)
	if err != nil {
		return bank, err
	}

	return bank, nil
}

func (s *Service) CreateBank(bank Bank) error {
	queryAccount := `
		SELECT
		    COALESCE(MAX(
		        RIGHT (accounts.code, 2)::INT), 0)
		FROM
		    accounts
		    LEFT JOIN accounts AS parent ON accounts.parent_id = parent.id
		WHERE
		    parent.code = $1`
	row := s.db.QueryRow(context.Background(), queryAccount, "1102")

	var maxNumber *int

	err := row.Scan(&maxNumber)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	id := uuid.New()

	var accountId string

	queryInsertToAccount := `
		INSERT INTO accounts (id, name, code, parent_id)
		SELECT
		    $1,
		    $2,
		    $3,
		    id
		FROM
		    accounts
		WHERE
		    code = '1102'
		RETURNING
		    id`

	err = tx.QueryRow(
		context.Background(),
		queryInsertToAccount,
		id,
		fmt.Sprintf(
			"%s-%s-%s",
			bank.Name.String,
			bank.Branch.String,
			bank.Number.String), fmt.Sprintf("1102%02d", *maxNumber+1)).Scan(&accountId)
	if err != nil {
		tx.Rollback(context.Background())

		return err
	}

	query := "INSERT INTO banks (id,name,branch,number,shaba,first_balance,type,account_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8);"

	bankId := uuid.New()

	_, err = tx.Exec(
		context.Background(),
		query,
		bankId,
		bank.Name,
		bank.Branch,
		bank.Number,
		bank.Shaba,
		bank.FirstBalance,
		bank.Type,
		accountId,
	)
	if err != nil {
		tx.Rollback(context.Background())

		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditBank(id string, bank Bank) error {
	query := "UPDATE banks SET name=$1,branch=$2,number=$3,shaba=$4,type=$5,first_balance=$6 WHERE id=$7;"
	validate := utils.NewValidate()

	err := validate.Struct(bank)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		context.Background(),
		query,
		bank.Name,
		bank.Branch,
		bank.Number,
		bank.Shaba,
		bank.Type,
		bank.FirstBalance,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteBank(id string, ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	defer tx.Rollback(ctx)

	if err != nil {
		return err
	}

	var accountId string

	query := "SELECT account_id FROM banks WHERE id=$1"

	err = tx.QueryRow(ctx, query, id).Scan(&accountId)
	if err != nil {
		return err
	}

	query = "DELETE FROM banks WHERE id=$1"

	_, err = tx.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	query = "DELETE FROM accounts WHERE id=$1"

	_, err = tx.Exec(ctx, query, accountId)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
