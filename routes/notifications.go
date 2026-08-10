package routes

import (
	"net/http"
	"strconv"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
)

func getAllNotification(context *gin.Context) {
	// Get all notification data from database by user's id
	notifs, err := models.GetNotificationsByRecipientID(context.GetInt64("uID"), 0)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, notifs)
}

func getNotificationWithLimit(context *gin.Context) {
	// Get limit from url parameter
	limit, err := strconv.Atoi(context.Param("limit"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Get limited notification data from database by user's id
	notifs, err := models.GetNotificationsByRecipientID(context.GetInt64("uID"), limit)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, notifs)
}

func markAllNotificationRead(context *gin.Context) {
	// Update is_read data by user's id
	err := models.UpdateIsReadByRecipientID(1, context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "All notifications are marked read"})
}

func markReadNotification(context *gin.Context) {
	// Get the notif id
	notifID, err := strconv.ParseInt(context.Param("notifID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Get the notification by id
	n, err := models.GetNotificationByID(notifID)

	// Update the notification's is_read value
	err = n.UpdateIsRead(1)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message":      "Notification is read",
		"notification": n,
	})
}
