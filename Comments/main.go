package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"math/rand"
	"net/http"
)

type Comment struct {
	Id     uint   `json:"id"`
	PostId uint   `json:"postId"`
	Text   string `json:"text"`
}

func main() {

	db, err := gorm.Open(mysql.Open("root:@tcp(127.0.0.1:3306)/comments_ms"), &gorm.Config{})

	if err != nil {
		panic(err)
	}
	db.AutoMigrate(Comment{})

	app := fiber.New()
	app.Use(cors.New())

	app.Get("/api/posts/:id/comments", func(ctx *fiber.Ctx) error {
		var comments []Comment
		db.Find(&comments, "post_id = ?", ctx.Params("id"))
		return ctx.JSON(comments)
	})

	app.Post("/api/comments", func(ctx *fiber.Ctx) error {
		var comment Comment
		if err := ctx.BodyParser(&comment); err != nil {
			return err
		}
		db.Create(&comment)
		// 10% chance of failure
		if rand.Intn(10) <= 5 {
			url := fmt.Sprintf("http://localhost:3000/api/posts/%d/comments", comment.PostId)
			body, _ := json.Marshal(map[string]string{"text": comment.Text})
			http.Post(url, "application/json", bytes.NewBuffer(body))
		}
		return ctx.JSON(comment)
	})

	app.Listen(":3001")
}
