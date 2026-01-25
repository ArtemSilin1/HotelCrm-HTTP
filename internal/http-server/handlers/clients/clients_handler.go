package clients_handler

import (
	"net/http"

	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/http-server/logger"
	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/storage/data"
	db_clients "github.com/ArtemSilin1/HotelCrm-HTTP/internal/storage/models/clients"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

const moduleName = "ClientsModule"

func NewHandler(db *pgxpool.Pool, logger *logger.Logger) *Handler {
	return &Handler{
		db:     db,
		logger: logger,
	}
}

type Handler struct {
	db     *pgxpool.Pool
	logger *logger.Logger
}

func (h *Handler) InitHandler(router *gin.Engine) {
	router.POST("/clients/add-client", h.AddClient)
	router.PUT("/clients/edit-client", h.EditClient)
	router.GET("/clients/get-clients-list", h.GetClients)
}

func (h *Handler) AddClient(c *gin.Context) {
	var clients db_clients.Clients

	if err := c.ShouldBindJSON(&clients); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": data.WrongData})
		return
	}

	if err := clients.AddClient(h.db); err != nil {
		logger.New("error", moduleName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "done"})
}

func (h *Handler) EditClient(c *gin.Context) {
	var clients db_clients.Clients
	if err := c.ShouldBindJSON(&clients); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": data.WrongData})
		return
	}

	if err := clients.EditClient(h.db); err != nil {
		logger.New("error", moduleName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "done"})
}

func (h *Handler) GetClients(c *gin.Context) {
	var clients db_clients.Clients

	clientsArr, err := clients.GetClients(h.db)
	if err != nil {
		logger.New("error", moduleName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": clientsArr})
}
