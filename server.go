package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo"
	"github.com/mallowww/assessment/expenses"
)

// ระบบสามารถจัดเก็บข้อมูล เรื่อง(title), ยอดค่าใช้จ่าย(amount), บันทึกย่อ(note) และ หมวดหมู่(tags)
// var db *sql.DB
var err error

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK")
}

func InitRoutes(e *echo.Echo) {
	e.GET("/expenses", expenses.GetExpensesHandler)
	e.GET("/expenses/:id", expenses.GetExpensesIdHandler)
	e.POST("/expenses", expenses.CreateExpensesHandler)
	e.PUT("/expenses/:id", expenses.UpdateExpensesHandler)
}

func main() {
	fmt.Println("Please use server.go for main file")
	fmt.Println("start at port:", os.Getenv("PORT"))
	e := echo.New()

	url := os.Getenv("DATABASE_URL")
	expenses.InitDB(url)

	// CORS ?
	// e.Use(middleware.CORS())

	// Middleware
	// e.Use(middleware.Logger())
	// e.Use(middleware.Recover())

	// Routing
	// InitRoutes()
	e.GET("/expenses", expenses.GetExpensesHandler)
	e.GET("/expenses/:id", expenses.GetExpensesIdHandler)
	e.POST("/expenses", expenses.CreateExpensesHandler)
	e.PUT("/expenses/:id", expenses.UpdateExpensesHandler)
	e.GET("/healthCheck", healthHandler)

	log.Println("server started at :2565")

	// Graceful Shut.
	go func() {
		err := e.Start(":2565")
		if err != nil {
			fmt.Println("server shutting down... - ", err)
		}
	}()
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = e.Shutdown(ctx)
	if err != nil {
		fmt.Println(err)
	}

}
