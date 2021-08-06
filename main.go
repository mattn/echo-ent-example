package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/mattn/echo-ent-example/ent"
	"github.com/mattn/echo-ent-example/ent/comment"
)

func setupEcho() *echo.Echo {
	e := echo.New()
	e.Debug = true
	e.Logger.SetOutput(os.Stderr)
	return e
}

// Error indicate response erorr
type Error struct {
	Error string `json:"error"`
}

// Controller is a controller for this application.
type Controller struct {
	client *ent.Client
}

// GetComment is GET handler to return record.
func (controller *Controller) GetComment(c echo.Context) error {
	// fetch record specified by parameter id
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Logger().Error("ParseInt: ", err)
		return c.String(http.StatusBadRequest, "ParseInt: "+err.Error())
	}
	comment, err := controller.client.Comment.Get(context.Background(), int(id))
	if err != nil {
		if !ent.IsNotFound(err) {
			c.Logger().Error("Get: ", err)
			return c.String(http.StatusBadRequest, "Get: "+err.Error())
		}
		return c.String(http.StatusNotFound, "Not Found")
	}
	return c.JSON(http.StatusOK, comment)
}

// ListComments is GET handler to return records.
func (controller *Controller) ListComments(c echo.Context) error {
	// fetch last 10 records
	cq := controller.client.Comment.Query().Order(ent.Desc(comment.FieldCreated)).Limit(10)
	comments, err := cq.All(context.Background())
	if err != nil {
		c.Logger().Error("All: ", err)
		return c.String(http.StatusBadRequest, "All: "+err.Error())
	}
	return c.JSON(http.StatusOK, comments)
}

// InsertComment is POST handler to insert record.
func (controller *Controller) InsertComment(c echo.Context) error {
	var comment ent.Comment
	// bind request to comment struct
	if err := c.Bind(&comment); err != nil {
		c.Logger().Error("Bind: ", err)
		return c.String(http.StatusBadRequest, "Bind: "+err.Error())
	}
	// insert record
	newComment := controller.client.Comment.Create()
	if comment.Name != "" {
		newComment.SetName(comment.Name)
	}
	newComment.SetText(comment.Text)
	if _, err := newComment.Save(context.Background()); err != nil {
		c.Logger().Error("Insert: ", err)
		return c.String(http.StatusBadRequest, "Save: "+err.Error())
	}
	c.Logger().Infof("inserted comment: %v", comment.ID)
	return c.NoContent(http.StatusCreated)
}

func main() {
	client, err := ent.Open("postgres", os.Getenv("DSN"))
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	controller := &Controller{client: client}

	e := setupEcho()

	e.GET("/api/comments/:id", controller.GetComment)
	e.GET("/api/comments", controller.ListComments)
	e.POST("/api/comments", controller.InsertComment)
	e.Static("/", "static/")
	e.Logger.Fatal(e.Start(":8989"))
}
