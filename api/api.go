package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type API struct {
	server *gin.Engine
	db     *pgxpool.Pool
}

func NewAPI(db *pgxpool.Pool) *API {
	return &API{
		db: db,
	}
}

func (api *API) CreateServer() *gin.Engine {
	r := gin.Default()

	r.GET("/health", Health)
	r.POST("/events", api.AddEvent)
	r.POST("/tasks", AddTask)
	r.GET("/events", api.ListEvents)
	return r
}
