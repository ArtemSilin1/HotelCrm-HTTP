package rooms

import (
	"net/http"

	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/http-server/logger"
	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/storage/data"
	db_rooms "github.com/ArtemSilin1/HotelCrm-HTTP/internal/storage/models/rooms"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

func NewHandler(db *pgxpool.Pool, logger *logger.Logger) *Handler {
	return &Handler{db: db, logger: logger}
}

type Handler struct {
	db     *pgxpool.Pool
	logger *logger.Logger
}

func (h *Handler) InitHandler(router *gin.Engine) {
	router.POST("/rooms/edit-room-status", h.EditStatus)
	router.GET("/rooms/get-rooms", h.GetRooms)
}

func (h *Handler) EditStatus(c *gin.Context) {
	var roomRequest db_rooms.Rooms
	if err := c.ShouldBindJSON(&roomRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response": data.WrongData})
		return
	}

	if err := roomRequest.EditRoomStatus(h.db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"response": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "done"})
}

func (h *Handler) GetRooms(c *gin.Context) {
	var roomRequest db_rooms.Rooms

	rooms, err := roomRequest.GetRooms(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"response": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": rooms})
}
