package store

import (
	"time"

	"github.com/lib/pq"
)

func (s *Store) SaveOperator(op *Operator) error {
	_, err := s.db.Exec(`
		INSERT INTO operators (id, name, license, email, password_hash)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING
	`, op.ID, op.Name, op.License, op.Email, op.PasswordHash)
	return err
}

func (s *Store) GetOperator(id string) (*Operator, bool) {
	op := &Operator{}
	err := s.db.QueryRow(`
		SELECT id, name, license, email, password_hash FROM operators WHERE id = $1
	`, id).Scan(&op.ID, &op.Name, &op.License, &op.Email, &op.PasswordHash)
	if err != nil {
		return nil, false
	}
	return op, true
}

func (s *Store) GetOperatorByEmail(email string) (*Operator, bool) {
	op := &Operator{}
	err := s.db.QueryRow(`
		SELECT id, name, license, email, password_hash FROM operators
		WHERE LOWER(email) = LOWER($1)
		ORDER BY created_at ASC
		LIMIT 1
	`, email).Scan(&op.ID, &op.Name, &op.License, &op.Email, &op.PasswordHash)
	if err != nil {
		return nil, false
	}
	return op, true
}

func (s *Store) SetOperatorPasswordHash(id, passwordHash string) bool {
	res, err := s.db.Exec(`UPDATE operators SET password_hash = $1 WHERE id = $2`, passwordHash, id)
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}

func (s *Store) SaveDrone(drone *Drone) error {
	status := drone.Status
	if status == "" {
		status = "available"
	}
	createdAt := drone.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	securityGoals := drone.SecurityGoals
	if securityGoals == nil {
		securityGoals = []string{}
	}

	_, err := s.db.Exec(`
		INSERT INTO drones (id, operator_id, name, security_goals, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`, drone.ID, drone.OperatorID, drone.Name, pq.Array(securityGoals), status, createdAt)
	return err
}

func (s *Store) ListDronesByOperator(operatorID string) []*Drone {
	rows, err := s.db.Query(`
		SELECT id, operator_id, name, security_goals, status, created_at
		FROM drones
		WHERE operator_id = $1
		ORDER BY created_at ASC
	`, operatorID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var drones []*Drone
	for rows.Next() {
		drone := &Drone{}
		var goals pq.StringArray
		if err := rows.Scan(&drone.ID, &drone.OperatorID, &drone.Name, &goals, &drone.Status, &drone.CreatedAt); err == nil {
			drone.SecurityGoals = []string(goals)
			drones = append(drones, drone)
		}
	}
	return drones
}

func (s *Store) FindExecutorDrone(requiredSecurityGoals []string) (*Drone, *Operator, bool) {
	if requiredSecurityGoals == nil {
		requiredSecurityGoals = []string{}
	}
	row := s.db.QueryRow(`
		SELECT
			d.id, d.operator_id, d.name, d.security_goals, d.status, d.created_at,
			o.id, o.name, o.license, o.email, o.password_hash
		FROM drones d
		JOIN operators o ON o.id = d.operator_id
		WHERE d.status = 'available'
			AND (cardinality($1::text[]) = 0 OR d.security_goals @> $1::text[])
		ORDER BY cardinality(d.security_goals) ASC, d.created_at ASC
		LIMIT 1
	`, pq.Array(requiredSecurityGoals))

	drone := &Drone{}
	operator := &Operator{}
	var goals pq.StringArray
	if err := row.Scan(
		&drone.ID,
		&drone.OperatorID,
		&drone.Name,
		&goals,
		&drone.Status,
		&drone.CreatedAt,
		&operator.ID,
		&operator.Name,
		&operator.License,
		&operator.Email,
		&operator.PasswordHash,
	); err != nil {
		return nil, nil, false
	}
	drone.SecurityGoals = []string(goals)
	return drone, operator, true
}
