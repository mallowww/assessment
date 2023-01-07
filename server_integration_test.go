package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func SetupTest(method, url string, body io.Reader, rec *httptest.ResponseRecorder) echo.Context {
	e := echo.New()
	req := httptest.NewRequest(method, url, body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	return c
}

func SetupNewDB(newDB *sql.DB) {
	db = newDB
}

func SetupDB(t *testing.T) sqlmock.Sqlmock {
	db, mock, errMock := sqlmock.New()
	if errMock != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", errMock)
	}
	SetupNewDB(db)

	return mock
}

// EXP01 - POST /expenses
func TestPostExpense(t *testing.T) {
	mock := SetupDB(t)

	var data = Expense{
		Title:  "strawberry smoothie",
		Amount: 79,
		Note:   "night market promotion discount 10 bath",
		Tags:   []string{"food", "beverage"},
	}

	mockedSql := "INSERT INTO expenses (title, amount, note, tags) values ($1,$2,$3,$4) RETURNING id, title, amount, note, tags"
	mockedRow := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta(mockedSql)).WithArgs(data.Title, data.Amount, data.Note, pq.Array(data.Tags)).WillReturnRows((mockedRow))

	body := bytes.NewBufferString(`{
		"title": "strawberry smoothie",
		"amount": 79.0,
		"note": "night market promotion discount 10 bath",
		"tags": ["food", "beverage"]
	}`)

	rec := httptest.NewRecorder()
	c := SetupTest(http.MethodPost, uri("expenses"), body, rec)
	errCreate := CreateExpensesHandler(c)
	assert.NoError(t, errCreate)

	var expense Expense
	json.NewDecoder(rec.Body).Decode(&expense)

	assert.EqualValues(t, http.StatusCreated, rec.Code)
	assert.NotEqual(t, data.ID, expense.ID)
	assert.Equal(t, data.Title, expense.Title)
	assert.Equal(t, data.Amount, expense.Amount)
	assert.Equal(t, data.Note, expense.Note)
	assert.Equal(t, data.Tags, expense.Tags)
}

// EXP02 - GET /expenses/:id
func TestGetExpenseId(t *testing.T) {
	mock := SetupDB(t)
	expensesMockRows := sqlmock.NewRows([]string{"id", "title", "amount", "note", "tags"}).
		AddRow("1", "strawberry smoothie", 79, "night market promotion discount 10 bath", pq.Array([]string{"food", "beverage"}))
	mockedSql := "SELECT id, title, amount, note, tags FROM expenses WHERE id = $1"
	mock.ExpectPrepare(regexp.QuoteMeta(mockedSql)).ExpectQuery().WithArgs(1).WillReturnRows(expensesMockRows)

	var latest Expense
	rec := httptest.NewRecorder()

	c := SetupTest(http.MethodGet, uri("expenses", "1"), nil, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := GetExpensesIdHandler(c)
	json.NewDecoder(rec.Body).Decode(&latest)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, latest.ID)
	assert.Equal(t, "strawberry smoothie", latest.Title)
	assert.Equal(t, 79, latest.Amount)
	assert.Equal(t, "night market promotion discount 10 bath", latest.Note)
	assert.Equal(t, []string{"food", "beverage"}, latest.Tags)
}

// EXP03 - PUT /expenses/:id
func TestPutExpenseId(t *testing.T) {
	mock := SetupDB(t)
	mockedSql := "UPDATE expenses SET title = $2, amount = $3, note = $4, tags = $5 WHERE id = $1"
	mockedRow := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectPrepare(regexp.QuoteMeta(mockedSql)).ExpectQuery().WithArgs(1, "apple smoothie", 89, "no discount", pq.Array([]string{"beverage"})).WillReturnRows((mockedRow))

	body := bytes.NewBufferString(`{
        "title": "apple smoothie",
        "amount": 89,
        "note": "no discount",
        "tags": ["beverage"]
    }`)
	var expense Expense
	rec := httptest.NewRecorder()
	c := SetupTest(http.MethodPut, uri("expenses", "1"), body, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := UpdateExpensesHandler(c)
	json.NewDecoder(rec.Body).Decode(&expense)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, expense.ID)
	assert.Equal(t, "apple smoothie", expense.Title)
	assert.Equal(t, 89, expense.Amount)
	assert.Equal(t, "no discount", expense.Note)
	assert.Equal(t, []string{"beverage"}, expense.Tags)
}

// EXP04 - GET /expenses
func TestGetAllExpense(t *testing.T) {
	mock := SetupDB(t)
	expensesMockRows := sqlmock.NewRows([]string{"id", "title", "amount", "note", "tags"}).
		AddRow("1", "apple smoothie", 89, "no discount", pq.Array([]string{"beverage"}))
	mock.ExpectPrepare("SELECT id, title, amount, note, tags FROM expenses").ExpectQuery().WillReturnRows(expensesMockRows)

	rec := httptest.NewRecorder()
	c := SetupTest(http.MethodGet, uri("expenses"), nil, rec)
	err := GetExpensesHandler(c)

	expected := "[{\"id\":1,\"title\":\"apple smoothie\",\"amount\":89,\"note\":\"no discount\",\"tags\":[\"beverage\"]}]"
	assert.Nil(t, err)
	assert.EqualValues(t, http.StatusOK, rec.Code)
	assert.Equal(t, expected, strings.TrimSpace(rec.Body.String()))
}

func uri(paths ...string) string {
	host := "http://localhost:2565"
	if paths == nil {
		return host
	}

	url := append([]string{host}, paths...)
	return strings.Join(url, "/")
}
