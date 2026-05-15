package store

import "github.com/lib/pq"

func (s *Store) SaveOrder(o *Order) error {
	securityGoals := o.SecurityGoals
	if securityGoals == nil {
		securityGoals = []string{}
	}

	_, err := s.db.Exec(`
		INSERT INTO orders (
			id, customer_id, description, budget,
			from_lat, from_lon, to_lat, to_lon,
			status, operator_id, offered_price,
			mission_type, security_goals,
			top_left_lat, top_left_lon, bottom_right_lat, bottom_right_lon,
			commission_amount, operator_amount,
			created_at
		)
		VALUES ($1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10, $11,
			$12, $13,
			$14, $15, $16, $17,
			$18, $19,
			$20)
		ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status
	`, o.ID, o.CustomerID, o.Description, o.Budget,
		o.FromLat, o.FromLon, o.ToLat, o.ToLon,
		string(o.Status), o.OperatorID, o.OfferedPrice,
		o.MissionType, pq.Array(securityGoals),
		o.TopLeftLat, o.TopLeftLon, o.BottomRightLat, o.BottomRightLon,
		o.CommissionAmount, o.OperatorAmount,
		o.CreatedAt)
	return err
}

func (s *Store) GetOrder(id string) (*Order, bool) {
	o := &Order{}
	err := s.db.QueryRow(`
		SELECT id, customer_id, description, budget,
			from_lat, from_lon, to_lat, to_lon,
			status, operator_id, offered_price,
			mission_type, security_goals,
			top_left_lat, top_left_lon, bottom_right_lat, bottom_right_lon,
			commission_amount, operator_amount,
			created_at
		FROM orders WHERE id = $1
	`, id).Scan(
		&o.ID, &o.CustomerID, &o.Description, &o.Budget,
		&o.FromLat, &o.FromLon, &o.ToLat, &o.ToLon,
		&o.Status, &o.OperatorID, &o.OfferedPrice,
		&o.MissionType, pq.Array(&o.SecurityGoals),
		&o.TopLeftLat, &o.TopLeftLon, &o.BottomRightLat, &o.BottomRightLon,
		&o.CommissionAmount, &o.OperatorAmount,
		&o.CreatedAt,
	)
	if err != nil {
		return nil, false
	}
	return o, true
}

func (s *Store) ListOrders() []*Order {
	rows, err := s.db.Query(`
		SELECT id, customer_id, description, budget,
			from_lat, from_lon, to_lat, to_lon,
			status, operator_id, offered_price,
			mission_type, security_goals,
			top_left_lat, top_left_lon, bottom_right_lat, bottom_right_lon,
			commission_amount, operator_amount,
			created_at
		FROM orders ORDER BY created_at DESC
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		o := &Order{}
		if err := rows.Scan(
			&o.ID, &o.CustomerID, &o.Description, &o.Budget,
			&o.FromLat, &o.FromLon, &o.ToLat, &o.ToLon,
			&o.Status, &o.OperatorID, &o.OfferedPrice,
			&o.MissionType, pq.Array(&o.SecurityGoals),
			&o.TopLeftLat, &o.TopLeftLon, &o.BottomRightLat, &o.BottomRightLon,
			&o.CommissionAmount, &o.OperatorAmount,
			&o.CreatedAt,
		); err != nil {
			continue
		}
		orders = append(orders, o)
	}
	return orders
}

func (s *Store) ListOrdersByCustomer(customerID string) []*Order {
	rows, err := s.db.Query(`
		SELECT id, customer_id, description, budget,
			from_lat, from_lon, to_lat, to_lon,
			status, operator_id, offered_price,
			mission_type, security_goals,
			top_left_lat, top_left_lon, bottom_right_lat, bottom_right_lon,
			commission_amount, operator_amount,
			created_at
		FROM orders WHERE customer_id = $1 ORDER BY created_at DESC
	`, customerID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		o := &Order{}
		if err := rows.Scan(
			&o.ID, &o.CustomerID, &o.Description, &o.Budget,
			&o.FromLat, &o.FromLon, &o.ToLat, &o.ToLon,
			&o.Status, &o.OperatorID, &o.OfferedPrice,
			&o.MissionType, pq.Array(&o.SecurityGoals),
			&o.TopLeftLat, &o.TopLeftLon, &o.BottomRightLat, &o.BottomRightLon,
			&o.CommissionAmount, &o.OperatorAmount,
			&o.CreatedAt,
		); err != nil {
			continue
		}
		orders = append(orders, o)
	}
	return orders
}

func (s *Store) UpdateOrderStatus(id string, status OrderStatus) bool {
	res, err := s.db.Exec(`UPDATE orders SET status = $1 WHERE id = $2`, string(status), id)
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}

func (s *Store) ConfirmPrice(id, operatorID string, acceptedPrice, commissionAmount float64) bool {
	if acceptedPrice <= 0 {
		return false
	}
	operatorAmount := acceptedPrice - commissionAmount
	res, err := s.db.Exec(`
		UPDATE orders 
		SET status = $1, offered_price = $2, commission_amount = $3, operator_amount = $4
		WHERE id = $5 AND status = $6 AND operator_id = $7
	`, string(StatusConfirmed), acceptedPrice, commissionAmount, operatorAmount, id, string(StatusMatched), operatorID)

	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}

func (s *Store) ConfirmCompletion(id string) bool {
	res, err := s.db.Exec(`
		UPDATE orders 
		SET status = $1 
		WHERE id = $2 AND status = $3
	`, string(StatusCompleted), id, string(StatusCompletedPending))

	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}

func (s *Store) SetOperatorOffer(orderID, operatorID string, price float64) bool {
	if price <= 0 {
		return false
	}
	res, err := s.db.Exec(`
		UPDATE orders 
		SET status = $1, operator_id = $2, offered_price = $3 
		WHERE id = $4 AND status IN ($5, $6)
	`, string(StatusMatched), operatorID, price, orderID, string(StatusSearching), string(StatusPending))

	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}

func (s *Store) ProcessOrderResult(orderID string, success bool) bool {
	targetStatus := StatusDispute
	if success {
		targetStatus = StatusCompletedPending
	}
	res, err := s.db.Exec(`
		UPDATE orders 
		SET status = $1 
		WHERE id = $2 AND status = $3
	`, string(targetStatus), orderID, string(StatusConfirmed))

	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}
