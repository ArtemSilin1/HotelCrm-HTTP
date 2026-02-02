package booking

import (
	"fmt"
	"net/http"

	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/storage/data"
	db_booking "github.com/ArtemSilin1/HotelCrm-HTTP/internal/storage/models/booking"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

func NewHandler(db *pgxpool.Pool) *HandlerBooking {
	return &HandlerBooking{db: db}
}

type HandlerBooking struct {
	db *pgxpool.Pool
}

func (h *HandlerBooking) InitHandler(router *gin.Engine) {
	router.POST("/booking/book", h.CreateBook)
	router.GET("/booking/book-list", h.GetBooks)
}

func (h *HandlerBooking) CreateBook(c *gin.Context) {
	var booking db_booking.Bookings

	if err := c.ShouldBindJSON(&booking); err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"response": data.WrongData})
	}

	if err := booking.Create(h.db); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"response": "done"})
}

func (h *HandlerBooking) GetBooks(c *gin.Context) {
	var booking db_booking.Bookings

	result, err := booking.Get(h.db)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"response": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": result})
}
