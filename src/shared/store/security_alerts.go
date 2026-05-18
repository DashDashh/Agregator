package store

import (
	"database/sql"

	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
)

type SecurityAlert = domain.SecurityAlert

func (s *Store) SaveSecurityAlert(alert *SecurityAlert) error {
	_, err := s.db.Exec(`
		INSERT INTO security_alerts (
			id, code, severity, source, order_id, message, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, alert.ID, alert.Code, alert.Severity, alert.Source, alert.OrderID, alert.Message, alert.Status, alert.CreatedAt)
	return err
}

func (s *Store) ListSecurityAlerts(status string, limit int) []*SecurityAlert {
	if limit <= 0 || limit > 200 {
		limit = 100
	}

	var rows *sql.Rows
	var err error
	if status == "" {
		rows, err = s.db.Query(`
			SELECT id, code, severity, source, order_id, message, status, created_at
			FROM security_alerts
			ORDER BY created_at DESC
			LIMIT $1
		`, limit)
	} else {
		rows, err = s.db.Query(`
			SELECT id, code, severity, source, order_id, message, status, created_at
			FROM security_alerts
			WHERE status = $1
			ORDER BY created_at DESC
			LIMIT $2
		`, status, limit)
	}
	if err != nil {
		return nil
	}
	defer rows.Close()

	var alerts []*SecurityAlert
	for rows.Next() {
		alert := &SecurityAlert{}
		if err := rows.Scan(
			&alert.ID,
			&alert.Code,
			&alert.Severity,
			&alert.Source,
			&alert.OrderID,
			&alert.Message,
			&alert.Status,
			&alert.CreatedAt,
		); err == nil {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

func (s *Store) ResolveSecurityAlert(id string) bool {
	res, err := s.db.Exec(`UPDATE security_alerts SET status = 'resolved' WHERE id = $1 AND status <> 'resolved'`, id)
	if err != nil {
		return false
	}
	n, err := res.RowsAffected()
	return err == nil && n > 0
}
