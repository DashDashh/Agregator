package store

func (s *Store) SaveCustomer(c *Customer) error {
	_, err := s.db.Exec(`
		INSERT INTO customers (id, name, email, phone, password_hash)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING
	`, c.ID, c.Name, c.Email, c.Phone, c.PasswordHash)
	return err
}

func (s *Store) GetCustomer(id string) (*Customer, bool) {
	c := &Customer{}
	err := s.db.QueryRow(`
		SELECT id, name, email, phone, password_hash FROM customers WHERE id = $1
	`, id).Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.PasswordHash)
	if err != nil {
		return nil, false
	}
	return c, true
}

func (s *Store) GetCustomerByEmail(email string) (*Customer, bool) {
	c := &Customer{}
	err := s.db.QueryRow(`
		SELECT id, name, email, phone, password_hash FROM customers
		WHERE LOWER(email) = LOWER($1)
		ORDER BY created_at ASC
		LIMIT 1
	`, email).Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.PasswordHash)
	if err != nil {
		return nil, false
	}
	return c, true
}

func (s *Store) SetCustomerPasswordHash(id, passwordHash string) bool {
	res, err := s.db.Exec(`UPDATE customers SET password_hash = $1 WHERE id = $2`, passwordHash, id)
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}
