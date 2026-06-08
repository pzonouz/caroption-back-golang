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

func (s *Service) ListPersonsWithSortFilterPagination(
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

	orderByClause := utils.BuildOrderBy("persons", sort, sortDirection)

	whereClause, args := utils.BuildWhere(
		"persons",
		filters,
		filterOperands,
		filterConditions,
		1,
	)
	query := fmt.Sprintf(`
		SELECT
		    persons.id,
		    persons.first_name,
		    persons.name,
		    persons.address,
		    persons.phone_number,
		    COALESCE(persons.debit,'0'),
		    COALESCE(persons.credit,'0'),
		    persons.created_at,
		    persons.updated_at
		FROM
		    persons
		%s 
		GROUP BY persons.name,persons.id,persons.first_name
		%s %s
		`, whereClause, orderByClause, pagedBy)

	rows, err := s.db.Query(context.Background(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer rows.Close()

	var persons []Person

	for rows.Next() {
		var person Person
		if err := rows.Scan(&person.ID, &person.FirstName, &person.Name, &person.Address, &person.PhoneNumber, &person.Debit, &person.Credit, &person.CreatedAt, &person.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		persons = append(persons, person)
	}

	w.Header().Add("Content-Type", "application/json")

	newQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM
		    persons
		%s
		`, whereClause)
	row := s.db.QueryRow(context.Background(), newQuery, args...)

	var Count int32

	err = row.Scan(&Count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var personsWithTotalCount struct {
		Rows       []Person `json:"rows"`
		TotalCount int32    `json:"totalCount"`
	}

	personsWithTotalCount.Rows = persons
	personsWithTotalCount.TotalCount = Count

	err = json.NewEncoder(w).Encode(personsWithTotalCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (s *Service) GetPerson(id string) (Person, error) {
	var person Person

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return person, err
	}

	query := "SELECT id,first_name,name,address,phone_number,debit,credit,created_at,updated_at FROM persons WHERE id=$1"
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(
		&person.ID,
		&person.FirstName,
		&person.Name,
		&person.Address,
		&person.PhoneNumber,
		&person.Debit,
		&person.Credit,
		&person.CreatedAt,
		&person.UpdatedAt,
	)
	if err != nil {
		return person, err
	}

	return person, nil
}

func (s *Service) CreatePerson(person Person) error {
	query := `
		INSERT INTO persons (id, first_name, name, address, phone_number,debit,credit)
		    VALUES ($1, $2, $3, $4, $5, $6, $7)`

	id := uuid.New()

	_, err := s.db.Exec(
		context.Background(),
		query,
		id,
		person.FirstName,
		person.Name,
		person.Address,
		person.PhoneNumber,
		person.Debit,
		person.Credit,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) EditPerson(id string, person Person) error {
	query := "UPDATE persons SET first_name=$1,name=$2,address=$3,phone_number=$4,debit=$5,credit=$6 WHERE id=$7;"

	_, err := s.db.Exec(
		context.Background(),
		query,
		person.FirstName,
		person.Name,
		person.Address,
		person.PhoneNumber,
		person.Debit,
		person.Credit,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeletePerson(id string) error {
	query := "DELETE FROM persons WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
