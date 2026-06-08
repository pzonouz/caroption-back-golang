package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (s *Service) CreateInvoiceProductsGroup(ipg InvoiceProductsGroup, ctx context.Context) error {

	invoiceProductsGroupId := uuid.New()
	query := `
		INSERT INTO invoice_products_groups (id, name, product_ids, created_at)
		    VALUES ($1, $2, $3, $4)`

	_, err := s.db.Exec(ctx, query,
		invoiceProductsGroupId,
		ipg.Name,
		ipg.ProductIds,
		time.Now(),
	)
	if err != nil {
		return err
	}
	return nil
}
func (s *Service) ListInvoiceProductsGroups() ([]InvoiceProductsGroup, error) {
	query := `
		SELECT
		    id,
		    name,
		    product_ids,
		    created_at
		FROM
		    invoice_products_groups`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []InvoiceProductsGroup{}, err
	}
	defer rows.Close()

	var invoiceProductsGroups []InvoiceProductsGroup

	for rows.Next() {
		var invoiceProductsGroup InvoiceProductsGroup
		if err := rows.Scan(&invoiceProductsGroup.ID, &invoiceProductsGroup.Name, &invoiceProductsGroup.ProductIds, &invoiceProductsGroup.CreatedAt); err != nil {
			return []InvoiceProductsGroup{}, err
		}

		invoiceProductsGroups = append(invoiceProductsGroups, invoiceProductsGroup)
	}

	return invoiceProductsGroups, nil
}

func (s *Service) GetInvoiceProductsGroup(id string) (InvoiceProductsGroup, error) {
	var invoiceProductsGroup InvoiceProductsGroup

	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return invoiceProductsGroup, err
	}

	query := `
		SELECT
		    id,
		    name,
		    product_ids,
		    created_at
		FROM
		    invoice_products_groups
		WHERE
		    id = $1`
	row := s.db.QueryRow(context.Background(), query, parsedUUID)

	err = row.Scan(&invoiceProductsGroup.ID, &invoiceProductsGroup.Name, &invoiceProductsGroup.ProductIds, &invoiceProductsGroup.CreatedAt)
	if err != nil {
		return invoiceProductsGroup, err
	}

	return invoiceProductsGroup, nil
}
func (s *Service) EditInvoiceProductsGroup(id string, invoiceProductsGroup InvoiceProductsGroup) error {
	query := `
		UPDATE
		    invoice_products_groups
		SET
		    name = $1,
		    product_ids = $2
		WHERE
		    id = $3`

	_, err := s.db.Exec(context.Background(), query, invoiceProductsGroup.Name, invoiceProductsGroup.ProductIds, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteInvoiceProductsGroup(id string) error {
	query := "DELETE FROM invoice_products_groups WHERE id=$1"

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
