package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/mattn/echo-ent-example/ent"
	_ "github.com/mattn/go-sqlite3"
)

func TestInsertCommentWithoutComment(t *testing.T) {
	client, err := ent.Open("sqlite3", ":memory:?_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()

	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	controller := &Controller{client: client}

	req := httptest.NewRequest(http.MethodPost, "/api/comments", strings.NewReader(`
	{
		"name": "job"
	}
	`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	e := setupEcho()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = controller.InsertComment(c)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ioutil.ReadAll(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	want := `validator failed for field "text": value is less than the required length`
	if !strings.Contains(got, want) {
		log.Fatalf("want %v but got %v", want, got)
	}
}

func TestInsertCommentWithComment(t *testing.T) {
	client, err := ent.Open("sqlite3", ":memory:?_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()

	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	controller := &Controller{client: client}

	req := httptest.NewRequest(http.MethodPost, "/api/comments", strings.NewReader(`
	{
		"name": "job",
		"text": "hello"
	}
	`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	e := setupEcho()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = controller.InsertComment(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != 201 {
		t.Fatal("should be succeeded")
	}
	if rec.Body.Len() > 0 {
		log.Fatal("response body should be empty")
	}
}

func TestGetComment(t *testing.T) {
	client, err := ent.Open("sqlite3", ":memory:?_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()

	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	controller := &Controller{client: client}

	req := httptest.NewRequest(http.MethodGet, "/api/comments/1", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	e := setupEcho()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/comments/:id")
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(1))

	err = controller.GetComment(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != 404 {
		t.Fatal("should be 404")
	}

	comment := ent.Comment{
		Text: "hello",
	}
	newComment := controller.client.Comment.Create()
	if comment.Name != "" {
		newComment.SetName(comment.Name)
	}
	newComment.SetText(comment.Text)
	if _, err := newComment.Save(context.Background()); err != nil {
		t.Fatal(err)
	}

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/api/comments/:id")
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(1))

	err = controller.GetComment(c)
	if err != nil {
		t.Fatal(err)
	}

	err = json.NewDecoder(rec.Body).Decode(&comment)
	if err != nil {
		t.Fatal(err)
	}
	want := "hello"
	got := comment.Text
	if got != want {
		log.Fatalf("want %v but got %v", want, got)
	}
}

func TestListComment(t *testing.T) {
	client, err := ent.Open("sqlite3", ":memory:?_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()

	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	controller := &Controller{client: client}

	req := httptest.NewRequest(http.MethodGet, "/api/comments", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	e := setupEcho()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/comments/:id")
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(1))

	err = controller.ListComments(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 {
		t.Fatal("should be 200")
	}
	var comments []ent.Comment
	err = json.NewDecoder(rec.Body).Decode(&comments)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) > 0 {
		t.Fatal("should be empty")
	}

	comment := ent.Comment{
		Text: "hello",
	}
	newComment := controller.client.Comment.Create()
	if comment.Name != "" {
		newComment.SetName(comment.Name)
	}
	newComment.SetText(comment.Text)
	if _, err := newComment.Save(context.Background()); err != nil {
		t.Fatal(err)
	}

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = controller.ListComments(c)
	if err != nil {
		t.Fatal(err)
	}

	err = json.NewDecoder(rec.Body).Decode(&comments)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) == 0 {
		t.Fatal("should not be empty")
	}
}
