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

	"github.com/lib/pq"
)

// ระบบสามารถจัดเก็บข้อมูล เรื่อง(title), ยอดค่าใช้จ่าย(amount), บันทึกย่อ(note) และ หมวดหมู่(tags)
type Expense struct {
	ID     int      `json:"id"`
	Title  string   `json:"title"`
	Amount float64  `json:"amount"`
	Note   string   `json:"note"`
	Tags   []string `json:"tags"`
}

type Err struct {
	Message string `json:"message"`
}

var db *sql.DB

// ดึงข้อมูลการใช้จ่ายทั้งหมด
func GetExpensesHandler(c echo.Context) error {
	var expenses = []Expense{}
	statement, err := db.Prepare("SELECT id, title, amount, note, tags FROM expenses")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't prepare query all expense statement"})
	}

	rows, err := statement.Query()
	if err != nil {
		// return c.JSON(http.StatusInternalServerError, Err{Message: "can't query all expense statement" + err.Error()})
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't query all expense statement" + err.Error()})
	}

	for rows.Next() {
		// e Expense struct
		var e Expense
		err = rows.Scan(&e.ID, &e.Title, &e.Amount, &e.Note, pq.Array(&e.Tags))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Err{Message: "can't scan expenses - " + err.Error()})
		}
		expenses = append(expenses, e)
	}
	return c.JSON(http.StatusOK, expenses)

}

// ดึงข้อมูลการใช้จ่ายทีละรายการ
func GetExpensesIdHandler(c echo.Context) error {
	id := c.Param("id")
	statement, err := db.Prepare("SELECT id, title, amount, note, tags FROM expenses WHERE id = $1")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't prepare expense by id" + err.Error()})
	}

	row := statement.QueryRow(id)
	e := Expense{}
	err = row.Scan(&e.ID, &e.Title, &e.Amount, &e.Note, pq.Array(&e.Tags))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, e)
}

// เพิ่มประวัติการใช้จ่ายใหม่ได้
func CreateExpensesHandler(c echo.Context) error {
	e := Expense{}
	err := c.Bind(&e)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}

	row := db.QueryRow("INSERT INTO expenses (title, amount, note, tags) values ($1,$2,$3,$4) RETURNING id, title, amount, note, tags", e.Title, e.Amount, e.Note, pq.Array(&e.Tags))
	err = row.Scan(&e.ID, &e.Title, &e.Amount, &e.Note, pq.Array(&e.Tags))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, e)

}

// ปรับเปลี่ยน/แก้ไข ข้อมูลของการใช้จ่ายได้
func UpdateExpensesHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK")

}

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK")
}

func main() {
	fmt.Println("Please use server.go for main file")
	fmt.Println("start at port:", os.Getenv("PORT"))
	e := echo.New()

	// InitDB()
	var err error
	url := os.Getenv("DATABASE_URL")
	db, err = sql.Open("postgres", url)
	if err != nil {
		log.Fatal("can't connect to database", err)
	}
	defer db.Close()

	// CreateTable()
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

	// CORS ?
	// e.Use(middleware.CORS())

	// Middleware
	// e.Use(middleware.Logger())
	// e.Use(middleware.Recover())


	// Routing
	e.GET("/expenses", GetExpensesHandler)
	e.GET("/expenses/:id", GetExpensesIdHandler)
	e.POST("/expenses", CreateExpensesHandler)
	e.PUT("/expenses/:id", UpdateExpensesHandler)
	e.GET("/healthCheck", healthHandler)

	log.Println("server started at :2565")
	// log.Fatal(e.Start(":2565"))

	go func() {
		err := e.Start(":2565")
		if err != nil {
			// fmt.Println("เซิฟปิดตัวลงด้วยเหตุผล - ", err)
			fmt.Println("server shutting down... - ", err)
		}
	}()
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// err := e.Shutdown(ctx)
	err = e.Shutdown(ctx)
	if err != nil {
		fmt.Println(err)
	}

}
