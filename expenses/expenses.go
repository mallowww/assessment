package expenses

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/lib/pq"
)

// Story: EXP01 - POST /expenses
// เพิ่มประวัติการใช้จ่ายใหม่ได้
func CreateExpensesHandler(c echo.Context) error {
	e := Expense{}
	err := c.Bind(&e)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "can't bind Expense{}" + err.Error()})
	}

	row := db.QueryRow("INSERT INTO expenses (title, amount, note, tags) values ($1,$2,$3,$4) RETURNING id, title, amount, note, tags", e.Title, e.Amount, e.Note, pq.Array(&e.Tags))
	err = row.Scan(&e.ID, &e.Title, &e.Amount, &e.Note, pq.Array(&e.Tags))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't create expense statement" + err.Error()})
	}
	return c.JSON(http.StatusCreated, e)
}

// Story: EXP02 - GET /expenses/:id
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

// Story: EXP03 - PUT /expenses/:id
// ปรับเปลี่ยน/แก้ไข ข้อมูลของการใช้จ่ายได้
func UpdateExpensesHandler(c echo.Context) error {
	e := Expense{}
	err := c.Bind(&e)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "can't bind Expense{}" + err.Error()})
	}

	id := c.Param("id")
	// statement, err := db.Prepare("UPDATE expenses SET title=$1, amount=$2, note=$3, tags=$4 WHERE id=$5")
	statement, err := db.Prepare("UPDATE expenses SET title=$2, amount=$3, note=$4, tags=$5 WHERE id=$1")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't prepare update expense statement" + err.Error()})
	}

	row, err := statement.Exec(&e.ID, &e.Title, &e.Amount, &e.Note, pq.Array(&e.Tags))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't execute statement" + err.Error()})
	}

	_, err = row.RowsAffected()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "expense rows doesn't affected row even after update statement" + err.Error()})
	}

	strId, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "can't convert id(int) to string" + err.Error()})
	}
	e.ID = strId
	return c.JSON(http.StatusOK, e)
}

// Story: EXP04 - GET /expenses
// ดึงข้อมูลการใช้จ่ายทั้งหมด
func GetExpensesHandler(c echo.Context) error {
	var expenses = []Expense{}
	statement, err := db.Prepare("SELECT id, title, amount, note, tags FROM expenses")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't prepare query all expense statement"})
	}

	rows, err := statement.Query()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't query all expense statement" + err.Error()})
	}

	for rows.Next() {
		var e Expense
		err = rows.Scan(&e.ID, &e.Title, &e.Amount, &e.Note, pq.Array(&e.Tags))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Err{Message: "can't scan expenses - " + err.Error()})
		}
		expenses = append(expenses, e)
	}
	return c.JSON(http.StatusOK, expenses)

}
