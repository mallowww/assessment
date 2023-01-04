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
		tags TEXT[]
		);`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal("can't create table", err)
	}
}

// Repository
func GetExpenses(db *sql.DB) ([]Expense, error) {
	// var expenses = []Expense{}
	expenses := []Expense{}
	statement, err := db.Prepare("SELECT id, title, amount, note, tags FROM expenses")
	if err != nil {
		// return c.JSON(http.StatusInternalServerError, Err{Message: "can't prepare query all expense statement"})
		return expenses, err
	}
	rows, err := statement.Query()
	if err != nil {
		// return c.JSON(http.StatusInternalServerError, Err{Message: "can't query all expense statement"})
		return expenses, err
	}

	for rows.Next() {
		var e Expense
		err = rows.Scan(&e.ID, &e.Title, &e.Amount, &e.Note, &e.Tags)
		if err != nil {
			// return c.JSON(http.StatusInternalServerError, Err{Message: "can't scan expenses - " + err.Error()})
			return expenses, err
		}
		expenses = append(expenses, e)
	}
	// return c.JSON(http.StatusOK, expenses)
	return expenses, nil
}

func CreateExpenses(db *sql.DB, e *Expense) error {
	// func CreateExpensesHandler(db *sql.DB) error {
	// var e Expense
	// err := c.Bind(&e)
	// if err != nil {
	// 	return c.JSON(http.StatusBadRequest, Err{Message: "can't bind to Expense - " + err.Error()})
	// }

	rows := db.QueryRow("INSERT INTO expenses (title, amount, note, tags) values ($1,$2,$3,$4) RETURNING id", &e.Title, &e.Amount, &e.Note, &e.Tags)
	err := rows.Scan(&e.ID)
	if err != nil {
		// return c.JSON(http.StatusBadRequest, Err{Message: "can't insert into expenses - " + err.Error()})
		return err
	}
	// return c.JSON(http.StatusCreated, "OK")
	return nil
}

// ดึงข้อมูลการใช้จ่ายทั้งหมด
func GetExpensesHandler(c echo.Context) error {
	expenses, err := GetExpenses(db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't prepare query all expense statement" + err.Error()})
	}
	return c.JSON(http.StatusOK, expenses)
}

// ดึงข้อมูลการใช้จ่ายทีละรายการ
func GetExpensesIdHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK")
}

// เพิ่มประวัติการใช้จ่ายใหม่ได้
func CreateExpensesHandler(c echo.Context) error {
	var e Expense
	err := c.Bind(&e)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "can't bind to Expense - " + err.Error()})
	}

	// err = rows.Scan(&e.ID)
	err = CreateExpenses(db, &e)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "can't insert into expenses - " + err.Error()})
	}
	return c.JSON(http.StatusCreated, "OK")
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
