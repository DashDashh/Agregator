package store

func (s *Store) RegisterIncident(i *Incident) error {
	_, err := s.db.Exec(`
		INSERT INTO incidents (
			id, order_id, operator_id, reporter_id,
			reason, description, damage_amount, status, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, i.ID, i.OrderID, i.OperatorID, i.ReporterID,
		i.Reason, i.Description, i.DamageAmount, i.Status, i.CreatedAt)
	if err != nil {
		return err
	}

	s.UpdateOrderStatus(i.OrderID, StatusDispute)
	return nil
}

func (s *Store) ListIncidentsByOrder(orderID string) []*Incident {
	rows, err := s.db.Query(`
		SELECT id, order_id, operator_id, reporter_id,
			reason, description, damage_amount, status, created_at
		FROM incidents WHERE order_id = $1 ORDER BY created_at DESC
	`, orderID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var incidents []*Incident
	for rows.Next() {
		i := &Incident{}
		if err := rows.Scan(
			&i.ID, &i.OrderID, &i.OperatorID, &i.ReporterID,
			&i.Reason, &i.Description, &i.DamageAmount, &i.Status, &i.CreatedAt,
		); err != nil {
			continue
		}
		incidents = append(incidents, i)
	}
	return incidents
}
