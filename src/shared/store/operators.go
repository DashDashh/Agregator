package store

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
