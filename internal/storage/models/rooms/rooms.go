package db_rooms

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	// Rooms statuses
	avalible    = "available"   // Доступный
	occupied    = "occupied"    // Занятый
	cleaning    = "cleaning"    // Уборка
	maintenance = "maintenance" // Обслуживание
)

type Rooms struct {
	Id            int     `json:"id"`
	RoomNumber    int     `json:"room_number"`
	RoomType      string  `json:"room_type"`
	PricePerNight float64 `json:"price_per_night"`
	BedroomsCount int     `json:"bedrooms_count"`
	Comment       string  `json:"comment"`
	Status        string  `json:"status"`
}

func (r *Rooms) checkRoomExist(db *pgxpool.Pool) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checkRoomExistQ := "SELECT COUNT(*) FROM rooms WHERE room_number = $1"

	var count int
	if err := db.QueryRow(ctx, checkRoomExistQ, r.RoomNumber).Scan(&count); err != nil {
		return false
	}

	return count > 0
}

func (r *Rooms) EditRoomStatus(db *pgxpool.Pool) error {
	isExist := r.checkRoomExist(db)
	if !isExist {
		return errors.New("room does not exist")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	editRoomStatusQ := "UPDATE rooms SET status = $1 WHERE room_nubmer = $2"

	_, err := db.Exec(ctx, editRoomStatusQ, r.Status, r.RoomNumber)
	if err != nil {
		return err
	}

	return nil
}

func (r *Rooms) GetRooms(db *pgxpool.Pool) ([]Rooms, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	GetClientsQ := "SELECT room_number, room_type, price_per_night, bedrooms_count, comment, status FROM rooms"
	rows, err := db.Query(ctx, GetClientsQ)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []Rooms
	for rows.Next() {
		var room Rooms
		if err := rows.Scan(
			&room.RoomNumber,
			&room.RoomType,
			room.PricePerNight,
			&room.BedroomsCount,
			&room.Comment,
			&room.Status); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}
