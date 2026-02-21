package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	transferRepo repository.InventoryTransferRepository
	productRepo  repository.ProductRepository
	branchRepo   repository.BranchRepository
}

func NewService(
	transferRepo repository.InventoryTransferRepository,
	productRepo repository.ProductRepository,
	branchRepo repository.BranchRepository,
) *Service {
	return &Service{
		transferRepo: transferRepo,
		productRepo:  productRepo,
		branchRepo:   branchRepo,
	}
}

type CreateTransferRequest struct {
	FromBranchID string   `json:"from_branch_id" binding:"required"`
	ToBranchID   string   `json:"to_branch_id" binding:"required"`
	ProductIDs   []string `json:"product_ids" binding:"required"`
	Notes        string   `json:"notes"`
	RequestedBy  string   `json:"requested_by" binding:"required"` // UserID
}

func (s *Service) CreateTransfer(ctx context.Context, req CreateTransferRequest) (*entity.InventoryTransfer, error) {
	fromBranchOID, err := primitive.ObjectIDFromHex(req.FromBranchID)
	if err != nil {
		return nil, errors.New("invalid from_branch_id")
	}
	toBranchOID, err := primitive.ObjectIDFromHex(req.ToBranchID)
	if err != nil {
		return nil, errors.New("invalid to_branch_id")
	}
	requestedByOID, err := primitive.ObjectIDFromHex(req.RequestedBy)
	if err != nil {
		return nil, errors.New("invalid requested_by")
	}

	// Verify branches exist
	if _, err := s.branchRepo.GetByID(ctx, fromBranchOID); err != nil {
		return nil, errors.New("source branch not found")
	}
	if _, err := s.branchRepo.GetByID(ctx, toBranchOID); err != nil {
		return nil, errors.New("destination branch not found")
	}

	// Generate Transfer Number
	// Assuming FromBranch is where the request originates or format "TR-YYYYMMDD-XXXX"
	transferNumber, err := s.transferRepo.GenerateTransferNumber(ctx)
	if err != nil {
		return nil, err
	}

	transfer := entity.NewInventoryTransfer(transferNumber, fromBranchOID, toBranchOID, requestedByOID)
	transfer.Notes = req.Notes

	// Process Items
	for _, pid := range req.ProductIDs {
		productID, err := primitive.ObjectIDFromHex(pid)
		if err != nil {
			return nil, fmt.Errorf("invalid product id: %s", pid)
		}

		// Verify product exists and belongs to FromBranch and is Available
		product, err := s.productRepo.GetByID(ctx, productID)
		if err != nil {
			return nil, fmt.Errorf("product not found: %s", pid)
		}
		if product.BranchID != fromBranchOID {
			return nil, fmt.Errorf("product %s does not belong to source branch", product.Name)
		}
		if product.Status != entity.ProductStatusAvailable {
			return nil, fmt.Errorf("product %s is not available (status: %s)", product.Name, product.Status)
		}

		// Add item with Quantity 1 (Serialized)
		transfer.AddItem(productID, 1)
	}

	if err := s.transferRepo.Create(ctx, transfer); err != nil {
		return nil, err
	}

	// Should we mark products as "InTransit" or "Reserved" immediately?
	// If allow validation, maybe "Reserved".
	// But Transfer has status Pending.
	// We should probably lock products so they aren't sold.
	// Update products status to Reserved?
	for _, item := range transfer.Items {
		product, _ := s.productRepo.GetByID(ctx, item.ProductID)
		if product != nil {
			product.Status = entity.ProductStatusReserved
			product.UpdatedAt = time.Now()
			_ = s.productRepo.Update(ctx, product) // Ignore error? Ideally transactional.
		}
	}

	return transfer, nil
}

func (s *Service) GetTransfers(ctx context.Context, branchID string, role string) ([]*entity.InventoryTransfer, error) {
	// If admin, show all? Or filter by branch.
	// For API simplicity, if branchID provided, filter.
	// Repo has GetByFromBranchID and GetByToBranchID.
	// We might need a generic Search or combine two queries.
	// Prioritize: If branchID is set, get transfers involving this branch (From or To).

	branchOID, err := primitive.ObjectIDFromHex(branchID)
	if err != nil {
		return nil, errors.New("invalid branch ID")
	}

	// This logic might be complex with current Repo interface (From/To separate).
	// For MVP, just return FromBranch transfers? Or need new Repo method `GetByBranchID`.
	// I'll stick to FromBranch for now or fetch both and merge.

	fromTransfers, _ := s.transferRepo.GetByFromBranchID(ctx, branchOID, nil)
	toTransfers, _ := s.transferRepo.GetByToBranchID(ctx, branchOID, nil)

	// Merge
	result := append(fromTransfers, toTransfers...)
	// Deduplicate? They shouldn't overlap unless From=To (impossible).

	return result, nil
}

func (s *Service) GetTransfer(ctx context.Context, id string) (*entity.InventoryTransfer, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID")
	}
	return s.transferRepo.GetByID(ctx, oid)
}

func (s *Service) ApproveTransfer(ctx context.Context, id, userID string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid ID")
	}
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	transfer, err := s.transferRepo.GetByID(ctx, oid)
	if err != nil {
		return err
	}

	if transfer.Status != entity.TransferStatusPending {
		return errors.New("transfer is not pending")
	}

	transfer.Approve(userOID)

	// Products remain Reserved (or change to InTransit if Product entity supported it? ProductStatusReserved is fine).

	return s.transferRepo.Update(ctx, transfer)
}

func (s *Service) ReceiveTransfer(ctx context.Context, id, userID string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid ID")
	}
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	transfer, err := s.transferRepo.GetByID(ctx, oid)
	if err != nil {
		return err
	}

	if transfer.Status != entity.TransferStatusInTransit {
		return errors.New("transfer is not in transit")
	}

	transfer.Receive(userOID)

	// Update Products: BranchID -> ToBranchID, Status -> Available
	for _, item := range transfer.Items {
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			continue // Log error?
		}
		product.BranchID = transfer.ToBranchID
		product.Status = entity.ProductStatusAvailable
		product.UpdatedAt = time.Now()
		if err := s.productRepo.Update(ctx, product); err != nil {
			return err
		}
	}

	return s.transferRepo.Update(ctx, transfer)
}

func (s *Service) CancelTransfer(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid ID")
	}

	transfer, err := s.transferRepo.GetByID(ctx, oid)
	if err != nil {
		return err
	}

	if transfer.Status != entity.TransferStatusPending {
		return errors.New("can only cancel pending transfers")
	}

	transfer.Cancel()

	// Revert Products to Available
	for _, item := range transfer.Items {
		product, _ := s.productRepo.GetByID(ctx, item.ProductID)
		if product != nil {
			product.Status = entity.ProductStatusAvailable
			product.UpdatedAt = time.Now()
			_ = s.productRepo.Update(ctx, product)
		}
	}

	return s.transferRepo.Update(ctx, transfer)
}
