package routes

import (
	"net/http"
	"strconv"

	"github.com/BatJoz21/cavejoz-go-api/models"
	"github.com/BatJoz21/cavejoz-go-api/utils"
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

	// Create friend request notification
	NotifyandPush(friendship.AddresseeID, friendship.RequesterID, friendship.ID, "friend_request")

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

	// Get friendship members
	idA, idB, err := models.GetFriendshipMembers(friendship.ID)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	// Delete conversation if exists
	err = models.DeleteConversationByMemberIDs(min(idA, idB), max(idA, idB))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Update stored friendship status to blocked at the database
	err = models.UpdateFriendshipStatusToBlocked(friendship.ID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User has been block"})
}

func getPendingFriendList(context *gin.Context) {
	// Get logged in user ID
	uID := context.GetInt64("uID")

	// Get pagination offset
	offset := utils.GetOffsetForPagination(models.FriendshipListDataMax, context)

	// Get pending status data from database
	pendings, err := models.GetPendingFriends(uID, offset)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Get total pending
	total := models.GetTotalPending(uID)

	context.JSON(http.StatusOK, gin.H{"data": pendings, "total": total})
}

func acceptFriendRequest(context *gin.Context) {
	// Get friendship data ID
	id, err := strconv.ParseInt(context.Param("frId"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Update stored friendship status to accepted at the database
	err = models.UpdateFriendshipStatusToAccepted(id, context.GetInt64("uID"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Sent notification to the requester
	reqID, err := models.GetRequesterID(id)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	NotifyandPush(reqID, context.GetInt64("uID"), id, "friend_accept")

	context.JSON(http.StatusOK, gin.H{
		"message": "Friend request accepted",
		"reqID":   reqID,
	})
}

func getFriendsList(context *gin.Context) {
	// Get logged in user ID
	uID := context.GetInt64("uID")

	// Get pagination offset
	offset := utils.GetOffsetForPagination(models.FriendshipListDataMax, context)

	// Get friends list
	friends, err := models.GetFriendshipList(uID, "accepted", offset)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Get total friends
	total, err := models.GetTotalFriendshipDataByUID(uID, "accepted")
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": friends, "total": total})
}

func getBlockedList(context *gin.Context) {
	// Get logged in user ID
	uID := context.GetInt64("uID")

	// Get pagination offset
	offset := utils.GetOffsetForPagination(models.FriendshipListDataMax, context)

	// Get blocked users list
	blocked, err := models.GetFriendshipList(uID, "blocked", offset)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Get total blocked
	total, err := models.GetTotalFriendshipDataByUID(uID, "blocked")
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": blocked, "total": total})
}

func getUserTotalFriend(context *gin.Context) {
	// Get user ID from parameter
	uID, err := strconv.ParseInt(context.Param("uID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Get total friends
	total, err := models.GetTotalFriendshipDataByUID(uID, "accepted")
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, total)
}

func getFriendshipStatus(context *gin.Context) {
	// Get user ID from parameter
	targetUID, err := strconv.ParseInt(context.Param("targetUID"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Get logged in user ID
	uID := context.GetInt64("uID")

	// Get Friendship status
	status := models.GetFriendshipStatus(uID, targetUID)

	context.JSON(http.StatusOK, status)
}

func deleteOrRejectFriendship(context *gin.Context) {
	// Get friendship data ID
	id, err := strconv.ParseInt(context.Param("frId"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Get friendship members
	idA, idB, err := models.GetFriendshipMembers(id)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	// Delete conversation if exists
	err = models.DeleteConversationByMemberIDs(min(idA, idB), max(idA, idB))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
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
