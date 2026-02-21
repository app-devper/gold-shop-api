package handler

import (
	"time"

	"github.com/devper-gold/gold-shop-api/app/feature/reward"
	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RewardHandler struct {
	service *reward.Service
}

func NewRewardHandler(service *reward.Service) *RewardHandler {
	return &RewardHandler{service: service}
}

func (h *RewardHandler) GetRewards(c *gin.Context) {
	onlyActive := c.Query("active") == "true"
	rewards, err := h.service.GetRewards(c.Request.Context(), onlyActive)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}
	utils.SuccessResponse(c, rewards)
}

func (h *RewardHandler) CreateReward(c *gin.Context) {
	var req struct {
		Code           string    `json:"code" binding:"required"`
		Name           string    `json:"name" binding:"required"`
		Description    string    `json:"description"`
		PointsRequired int       `json:"points_required" binding:"required"`
		Quantity       int       `json:"quantity" binding:"required"`
		ValidFrom      time.Time `json:"valid_from"`
		ValidUntil     time.Time `json:"valid_until"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	reward := entity.NewReward(req.Code, req.Name, req.Description, req.PointsRequired, req.Quantity, req.ValidFrom, req.ValidUntil)
	if err := h.service.CreateReward(c.Request.Context(), reward); err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.CreatedResponse(c, reward)
}

func (h *RewardHandler) RedeemReward(c *gin.Context) {
	var req struct {
		CustomerID string `json:"customer_id" binding:"required"`
		RewardID   string `json:"reward_id" binding:"required"`
		BranchID   string `json:"branch_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	custOID, _ := primitive.ObjectIDFromHex(req.CustomerID)
	rewOID, _ := primitive.ObjectIDFromHex(req.RewardID)
	brOID, _ := primitive.ObjectIDFromHex(req.BranchID)
	processedBy, _ := primitive.ObjectIDFromHex(c.GetString("userID"))

	redemption, err := h.service.RedeemReward(c.Request.Context(), custOID, rewOID, brOID, processedBy)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.CreatedResponse(c, redemption)
}
