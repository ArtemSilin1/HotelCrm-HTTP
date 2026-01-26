package db_rooms

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	// Rooms statuses
	available   = "available"   // Доступный
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

	editRoomStatusQ := "UPDATE rooms SET status = $1 WHERE room_number = $2"

	_, err := db.Exec(ctx, editRoomStatusQ, r.Status, r.RoomNumber)
	if err != nil {
		fmt.Println(err.Error())
		return fmt.Errorf("failed to update room status: %w", err)
	}

	return nil
}

func (r *Rooms) GetRooms(db *pgxpool.Pool) ([]Rooms, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Исправлено имя запроса (было GetClientsQ, должно быть GetRoomsQ)
	GetRoomsQ := "SELECT id, room_number, room_type, price_per_night, bedrooms_count, comment, status FROM rooms"
	rows, err := db.Query(ctx, GetRoomsQ)
	if err != nil {
		return nil, fmt.Errorf("failed to query rooms: %w", err)
	}
	defer rows.Close()

	var rooms []Rooms
	for rows.Next() {
		var room Rooms
		// Исправлено: добавлен & для PricePerNight и добавлено поле Id
		if err := rows.Scan(
			&room.Id,
			&room.RoomNumber,
			&room.RoomType,
			&room.PricePerNight, // <- ИСПРАВЛЕНО: добавлен &
			&room.BedroomsCount,
			&room.Comment,
			&room.Status); err != nil {
			return nil, fmt.Errorf("failed to scan room row: %w", err)
		}
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return rooms, nil
}

// Дополнительные методы для работы с decimal

// Метод для безопасного получения комнаты по ID
func GetRoomByID(db *pgxpool.Pool, id int) (*Rooms, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := "SELECT id, room_number, room_type, price_per_night, bedrooms_count, comment, status FROM rooms WHERE id = $1"

	var room Rooms
	err := db.QueryRow(ctx, query, id).Scan(
		&room.Id,
		&room.RoomNumber,
		&room.RoomType,
		&room.PricePerNight,
		&room.BedroomsCount,
		&room.Comment,
		&room.Status)

	if err != nil {
		return nil, fmt.Errorf("failed to get room by id %d: %w", id, err)
	}

	return &room, nil
}
