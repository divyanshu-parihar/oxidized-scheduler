package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/divyanshu-parihar/oxidized-scheduler/cmd/wheel"
)

type API struct {
	server *gin.Engine
	db     *pgxpool.Pool
	wheel  *wheel.Wheel
}

func NewAPI(db *pgxpool.Pool, w *wheel.Wheel) *API {
	return &API{
		db:    db,
		wheel: w,
	}
}

func (api *API) CreateServer() *gin.Engine {
	r := gin.Default()

	r.GET("/health", Health)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/wheel/state", api.GetWheelState)
	r.POST("/events", api.AddEvent)
	r.POST("/tasks", AddTask)
	r.PUT("/tasks/:id/status", api.UpdateTaskStatus)
	r.GET("/events", api.ListEvents)
	return r
}
