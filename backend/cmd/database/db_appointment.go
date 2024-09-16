package database

type Appointment struct {
	User             string `json:"user"`
	Merchant         string `json:"merchant"`
	Appointment_type string `json:"type"`
	Location         string `json:"location"`
	From_date        string `json:"from_date"`
	To_date          string `json:"to_date"`
}

func (s *service) NewAppointment(app Appointment) error {
	query := `
	insert into appointment("user", merchant, type, location, from_date, to_date)
	values ($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.Exec(query, app.User, app.Merchant, app.Appointment_type, app.Location, app.From_date, app.To_date)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) GetAppointmentsByUser(user string) ([]Appointment, error) {
	query := `
	select * from appointment where "user" = $1
	`

	rows, err := s.db.Query(query, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []Appointment
	for rows.Next() {
		var app Appointment
		if err := rows.Scan(&app.User, &app.Merchant, &app.Appointment_type, &app.Location, &app.From_date, &app.To_date); err != nil {
			return nil, err
		}
		appointments = append(appointments, app)
	}

	return appointments, nil
}

func (s *service) GetAppointmentsByMerchant(merchant string) ([]Appointment, error) {
	query := `
	select * from appointment where merchant = $1
	`

	rows, err := s.db.Query(query, merchant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []Appointment

	for rows.Next() {
		var app Appointment
		if err := rows.Scan(&app.User, &app.Merchant, &app.Appointment_type, &app.Location, &app.From_date, &app.To_date); err != nil {
			return nil, err
		}
		appointments = append(appointments, app)
	}

	return appointments, nil
}
