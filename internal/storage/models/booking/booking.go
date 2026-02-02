package db_booking

import (
	"context"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Bookings struct {
	Id         int                `json:"id"`
	ClientId   int                `json:"client_id"`
	RoomId     int                `json:"room_id"`
	Checkin    pgtype.Date        `json:"check_in_data"`
	Checkout   pgtype.Date        `json:"check_out_data"`
	TotalPrice float64            `json:"total_price"`
	Notes      string             `json:"notes"`
	CreatedAt  pgtype.Timestamptz `json:"created_at"`
}

func (b *Bookings) Create(db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	createQ :=
		`INSERT INTO Bookings(client_id, room_id, check_in_date, check_out_date, total_price, notes)
		VALUES($1, $2, $3, $4, $5, $6)
	`

	_, err := db.Exec(ctx, createQ)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bookings) Get(db *pgxpool.Pool) ([]Bookings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	getBookingQ := "SELECT id, client_id, room_id, check_in_date, check_out_date, total_price, notes FROM Bookings"
	rows, err := db.Query(ctx, getBookingQ)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bookings []Bookings
	for rows.Next() {
		var booking Bookings
		if err := rows.Scan(
			&booking.Id,
			&booking.ClientId,
			&booking.RoomId,
			&booking.Checkin,
			&booking.Checkout,
			&booking.TotalPrice,
		); err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil
}
