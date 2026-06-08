package services

import "context"

func (s *Service) ListSettings() ([]Setting, error) {
	query := `SELECT id,key,value FROM settings;`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return []Setting{}, err
	}

	var settings []Setting

	defer rows.Close()

	for rows.Next() {
		var row Setting
		rows.Scan(&row.ID, &row.Key, &row.Value)
		settings = append(settings, row)
	}

	return settings, nil
}
