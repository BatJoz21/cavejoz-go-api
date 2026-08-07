package routes

import (
	"net/http"
	"strconv"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
)

func getAllNotification(context *gin.Context) {
	notifs, err := models.GetNotificationsByRecipientID(context.GetInt64("uID"), 0)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, notifs)
}

func getNotificationWithLimit(context *gin.Context) {
	limit, err := strconv.Atoi(context.Param("limit"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	notifs, err := models.GetNotificationsByRecipientID(context.GetInt64("uID"), limit)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, notifs)
}
