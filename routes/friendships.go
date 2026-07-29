package routes

import (
	"net/http"
	"strconv"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/gin-gonic/gin"
)

func addFriend(context *gin.Context) {
	// Get addressee ID
	var dto models.AddFriendDTO
	err := context.ShouldBindBodyWithJSON(&dto)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	addresseeID, err := strconv.ParseInt(dto.AddresseeID, 10, 64)

	// Create Friendship struct
	friendship := models.Friendship{
		RequesterID: context.GetInt64("uID"),
		AddresseeID: addresseeID,
		Status:      "pending",
	}

	// Check if friendship status is pending, accepted, or blocked
	isExists, _ := models.IsFriendshipDataExists(friendship.RequesterID, friendship.AddresseeID)
	if isExists {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Failed to add this user as a friend"})
		return
	}

	// Insert new friendship data to database
	err = friendship.SaveAddFriendData()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Friend request sent"})
}

func blockAUser(context *gin.Context) {
	// Get addressee ID
	var dto models.AddFriendDTO
	err := context.ShouldBindBodyWithJSON(&dto)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	addresseeID, err := strconv.ParseInt(dto.AddresseeID, 10, 64)

	// Create Friendship struct
	friendship := models.Friendship{
		RequesterID: context.GetInt64("uID"),
		AddresseeID: addresseeID,
		Status:      "pending",
	}

	// Check if friendship data exists
	isExists, id := models.IsFriendshipDataExists(friendship.RequesterID, friendship.AddresseeID)
	// If not, create new friendship data
	if !isExists {
		// Insert new friendship data to database
		err = friendship.SaveAddFriendData()
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	} else {
		// If exists get the friendship data ID
		friendship.ID = id
	}

	// Update stored friendship status to blocked at the database
	err = models.UpdateFriendshipStatus("blocked", friendship.ID, friendship.AddresseeID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User has been block"})
}

func getPendingFriendList(context *gin.Context) {
	// Get logged in user ID
	uID := context.GetInt64("uID")

	// Get pending status data from database
	pendings, err := models.GetPendingFriends(uID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, pendings)
}

func acceptFriendRequest(context *gin.Context) {
	// Get friendship data ID
	id, err := strconv.ParseInt(context.Param("frId"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Update stored friendship status to accepted at the database
	err = models.UpdateFriendshipStatus("accepted", id, context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Friend request accepted"})
}

func getFriendsList(context *gin.Context) {
	// Get logged in user ID
	uID := context.GetInt64("uID")

	// Get friends list
	friends, err := models.GetUserFriends(uID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, friends)
}

func deleteOrRejectFriendship(context *gin.Context) {
	// Get friendship data ID
	id, err := strconv.ParseInt(context.Param("frId"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Remove friendship data from database
	err = models.DeleteFriendship(id)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Friend deleted"})
}
