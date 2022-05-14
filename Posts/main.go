package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type Post struct {
	Id           uint                `json:"id"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	Comments     []map[string]string `json:"comments" gorm:"-" default:"[]"`
	CommentsJson string              `json:"-"`
}

type Comment struct {
	//Id     uint   `json:"id"`
	PostId uint   `json:"postId"`
	Text   string `json:"text"`
}

func main() {
	db, err := gorm.Open(mysql.Open("root:@tcp(127.0.0.1:3306)/posts_ms"), &gorm.Config{})

	if err != nil {
		panic(err)
	}
	db.AutoMigrate(Post{})

	app := fiber.New()

	app.Use(cors.New())
	app.Get("/api/posts", func(ctx *fiber.Ctx) error {
		var posts []Post
		db.Find(&posts)
		for i, post := range posts {

			var comments []map[string]string
			json.Unmarshal([]byte(post.CommentsJson), &comments)
			posts[i].Comments = comments
		}
		return ctx.JSON(posts)

	})

	app.Post("api/post", func(ctx *fiber.Ctx) error {
		var post Post
		if err := ctx.BodyParser(&post); err != nil {
			return err
		}

		db.Create(&post)
		return ctx.JSON(post)
	})

	app.Post("/api/posts/:id/comments", func(ctx *fiber.Ctx) error {
		id, _ := strconv.Atoi(ctx.Params("id"))

		post := Post{Id: uint(id)}
		db.Model(Post{}).First(&post)
		var body map[string]string
		if err := ctx.BodyParser(&body); err != nil {
			return err
		}
		var comments []map[string]string
		json.Unmarshal([]byte(post.CommentsJson), &comments)
		comments = append(comments, map[string]string{
			"text": body["text"],
		})

		commentsJson, _ := json.Marshal(comments)
		db.Model(post).Where("id = ?", id).Update("comments_json", commentsJson)
		return ctx.JSON(post)
	})

	go Cron()
	app.Listen(":3000")
}

func Cron() {
	postsDB, err := gorm.Open(mysql.Open("root:@tcp(127.0.0.1:3306)/posts_ms"), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	commentsDB, err := gorm.Open(mysql.Open("root:@tcp(127.0.0.1:3306)/comments_ms"), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(time.Minute)

		var comments []Comment
		commentsDB.Find(&comments)
		var posts []Post
		postsDB.Find(&posts)

		for _, post := range posts {
			var postComments []map[string]string
			json.Unmarshal([]byte(post.CommentsJson), &postComments)
			fmt.Println(&postComments)

			var filteredComments []map[string]string
			for _, comment := range comments {
				if comment.PostId == post.Id {
					filteredComments = append(filteredComments, map[string]string{"text": comment.Text})
				}
			}

			if len(postComments) < len(filteredComments) {
				commentsJson, _ := json.Marshal(filteredComments)
				postsDB.Model(post).Where("id = ?", post.Id).Update("comments_json", commentsJson)
			}
		}

	}
}
