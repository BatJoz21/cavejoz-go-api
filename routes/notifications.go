package routes

import (
	"log"
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

func NotifyandPush(recipientID, actorID, refID int64, notifType string) {
	if recipientID == actorID {
		return
	}

	notif := models.Notification{
		RecipientID: recipientID,
		ActorID:     actorID,
		Type:        notifType,
		ReferenceID: refID,
		IsRead:      false,
	}
	if err := notif.Save(); err != nil {
		log.Printf("notification insert failed: %v", err)
		return
	}

	viewNotif, err := models.GetViewNotificationByID(notif.ID)
	if err != nil {
		log.Printf("notification fetch failed: %v", err)
		return
	}

	appHub.Send(recipientID, gin.H{
		"type":         "notification",
		"notification": viewNotif,
	})
}
