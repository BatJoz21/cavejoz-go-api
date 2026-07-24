package routes

import (
	"net/http"
	"strconv"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
)

func addFriend(context *gin.Context) {
	var dto models.AddFriendDTO
	err := context.ShouldBindBodyWithJSON(&dto)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	addresseeID, err := strconv.ParseInt(dto.AddresseeID, 10, 64)

	friendship := models.Friendship{
		RequesterID: context.GetInt64("uID"),
		AddresseeID: addresseeID,
		Status:      "pending",
	}
	err = friendship.SaveAddFriendData()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Friend request sent"})
}

func acceptFriendRequest(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err = models.UpdateFriendshipStatus("accepted", id, context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Friend request accepted"})
}

func getPendingFriendList(context *gin.Context) {
	uID := context.GetInt64("uID")

	pendings, err := models.GetPendingFriends(uID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, pendings)
}

func getFriendsList(context *gin.Context) {
	uID := context.GetInt64("uID")

	friends, err := models.GetUserFriends(uID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, friends)
}
