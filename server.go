package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
)

// ระบบสามารถจัดเก็บข้อมูล เรื่อง(title), ยอดค่าใช้จ่าย(amount), บันทึกย่อ(note) และ หมวดหมู่(tags)
type Expense struct {
	ID     int
	Title  string   `json:"เรื่อง"`
	Amount float64  `json:"ยอดค่าใช้จ่าย"`
	Note   string   `json:"บันทึกย่อ"`
	Tags   []string `json:"หมวดหมู่"`
}

func main() {
	e := echo.New()

	e.GET("/nanglen", func(c echo.Context) error {
		// time.Sleep(5 * time.Second)
		return c.JSON(200, nil)
	})

	// เอามาดู cli สั่งปิด server
	e.GET("/stopServer", stopServer)

	go func() {
		// err := e.Logger.Fatal(e.Start(":2565"))
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

func stopServer(c echo.Context) error {
	err := c.Echo().Shutdown(context.Background())
	if err != nil || err == http.ErrServerClosed {
		c.Echo().Logger.Fatal("shutting down this server")
	}
	return nil
}
