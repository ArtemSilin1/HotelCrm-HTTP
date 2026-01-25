package db_clients

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/storage/data"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Clients struct {
	Id        int    `json:"id"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Notes     string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

func (c *Clients) checkClientExist(db *pgxpool.Pool) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checkClientQ := "SELECT COUNT(*) FROM Clients WHERE id = $1"

	var count int
	if err := db.QueryRow(ctx, checkClientQ, c.Id).Scan(&count); err != nil {
		return false
	}

	return count > 0
}

func (c *Clients) AddClient(db *pgxpool.Pool) error {
	isExist := c.checkClientExist(db)
	if isExist {
		return fmt.Errorf(data.UserExists)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addClientQ := "INSERT INTO Clients (full_name, email, phone, notes) VALUES ($1, $2, $3, $4)"

	_, err := db.Exec(ctx, addClientQ, c.FullName, c.Email, c.Phone, c.Notes)
	if err != nil {
		return err
	}

	return nil
}

func (c *Clients) EditClient(db *pgxpool.Pool) error {
	isExist := c.checkClientExist(db)
	if !isExist {
		return errors.New("client not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	EditUserQ := `
					UPDATE Clients
					SET full_name = $1, email = $2, phone = $3, notes = $4
					WHERE id = $5
	`

	_, err := db.Exec(ctx, EditUserQ, c.FullName, c.Email, c.Phone, c.Notes, c.Id)
	if err != nil {
		return err
	}

	return nil
}

func (c *Clients) GetClients(db *pgxpool.Pool) ([]Clients, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	GetClientsQ := "SELECT id, full_name, email, phone, notes FROM Clients"

	rows, err := db.Query(ctx, GetClientsQ)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []Clients
	for rows.Next() {
		var client Clients
		if err := rows.Scan(&client.Id, &client.FullName, &client.Email, &client.Phone, &client.Notes); err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clients, nil
}
