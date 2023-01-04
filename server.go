package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "github.com/lib/pq"
)

// ระบบสามารถจัดเก็บข้อมูล เรื่อง(title), ยอดค่าใช้จ่าย(amount), บันทึกย่อ(note) และ หมวดหมู่(tags)
type Expense struct {
	ID     int
	Title  string   `json:"เรื่อง"`
	Amount float64  `json:"ยอดค่าใช้จ่าย"`
	Note   string   `json:"บันทึกย่อ"`
	Tags   []string `json:"หมวดหมู่"`
}

// ไว้เก็บ err msg
type Err struct {
	Message string `json:"message"`
}

// type CustomerHandler struct {
// 	DB *sql.DB
// }

var db *sql.DB

// var expenses = []Expense{}

func InitDB() {
	url := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatal("can't connect to database", err)
	}
	defer db.Close()

	// convert go struct to table - https://cheikhshift.github.io/struct-to-sql/
	createTable := `CREATE TABLE IF NOT EXISTS expenses (
		id SERIAL PRIMARY KEY,
		title TEXT, 
		amount FLOAT, 
		note TEXT, 
		tags VARCHAR
		);`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal("can't create table", err)
	}
}

// ดึงข้อมูลการใช้จ่ายทั้งหมด
func GetExpensesHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK")

}

// ดึงข้อมูลการใช้จ่ายทีละรายการ
func GetExpensesIdHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK")

}

// เพิ่มประวัติการใช้จ่ายใหม่ได้
func CreateExpensesHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK")

}

// ปรับเปลี่ยน/แก้ไข ข้อมูลของการใช้จ่ายได้
func UpdateExpensesHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK")

}

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK")
}

func main() {
	e := echo.New()
	InitDB()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Init handler
	// h := CustomerHandler{}
	// h.Initialize()

	// Routing
	e.GET("/expenses", GetExpensesHandler)
	e.GET("/expenses/:id", GetExpensesIdHandler)
	e.POST("/expenses", CreateExpensesHandler)
	e.PUT("/expenses/:id", UpdateExpensesHandler)
	e.GET("/healthCheck", healthHandler)

	go func() {
		err := e.Start(":2565")
		if err != nil {
			fmt.Println("เซิฟปิดตัวลงด้วยเหตุผล - ", err)
		}
	}()
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err := e.Shutdown(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
