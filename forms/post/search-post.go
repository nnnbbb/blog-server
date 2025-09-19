package post

import "blog-server/models"

type SearchPost struct {
	models.Post
	Score float64
}
