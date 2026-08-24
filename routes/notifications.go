package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
)

func getAllNotification(context *gin.Context) {
	// Get all notification data from database by user's id
	notifs, err := models.GetNotificationsByRecipientID(context.GetInt64("uID"), 0)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch notifications"})
		return
	}

	context.JSON(http.StatusOK, notifs)
}

func getNotificationWithLimit(context *gin.Context) {
	// Get limit from url parameter
	limit, err := strconv.Atoi(context.Param("limit"))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid limit"})
		return
	}

	// Get limited notification data from database by user's id
	notifs, err := models.GetNotificationsByRecipientID(context.GetInt64("uID"), limit)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch notifications"})
		return
	}

	context.JSON(http.StatusOK, notifs)
}

func markAllNotificationRead(context *gin.Context) {
	// Update is_read data by user's id
	err := models.UpdateIsReadByRecipientID(1, context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to mark all notifications read"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "All notifications are marked read"})
}

func markReadNotification(context *gin.Context) {
	// Get the notif id
	notifID, err := strconv.ParseInt(context.Param("notifID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid notifID"})
		return
	}

	// Get the notification by id
	n, err := models.GetNotificationByID(notifID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			context.JSON(http.StatusNotFound, gin.H{"message": "Notification not found"})
			return
		} else {
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch notification"})
			return
		}
	}

	// Update the notification's is_read value
	err = n.UpdateIsRead(1)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to mark notification read"})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message":      "Notification is read",
		"notification": n,
	})
}

func NotifyandPush(recipientID, actorID, refID int64, notifType string) {
	// Check if actor and recipient is the same user
	if recipientID == actorID {
		return
	}

	// Save notification data in the database
	notif := models.Notification{
		RecipientID: recipientID,
		ActorID:     actorID,
		Type:        notifType,
		ReferenceID: refID,
		IsRead:      false,
	}
	if err := notif.Save(); err != nil {
		return
	}

	// Get the notification view data from database
	viewNotif, err := models.GetViewNotificationByID(notif.ID)
	if err != nil {
		return
	}

	// Using web socket send it in JSON format
	appHub.Send(recipientID, gin.H{
		"type":         "notification",
		"notification": viewNotif,
	})
}
