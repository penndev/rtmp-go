package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Handleflv(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "haojiahuo",
	})
}
