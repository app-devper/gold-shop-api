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
	txManager    repository.TransactionManager
}

func NewService(
	transferRepo repository.InventoryTransferRepository,
	productRepo repository.ProductRepository,
	branchRepo repository.BranchRepository,
	txManager repository.TransactionManager,
) *Service {
	return &Service{
		transferRepo: transferRepo,
		productRepo:  productRepo,
		branchRepo:   branchRepo,
		txManager:    txManager,
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
	if fromBranchOID == toBranchOID {
		return nil, errors.New("from_branch and to_branch must differ")
	}
	requestedByOID, err := primitive.ObjectIDFromHex(req.RequestedBy)
	if err != nil {
		return nil, errors.New("invalid requested_by")
	}

	if _, err := s.branchRepo.GetByID(ctx, fromBranchOID); err != nil {
		return nil, errors.New("source branch not found")
	}
	if _, err := s.branchRepo.GetByID(ctx, toBranchOID); err != nil {
		return nil, errors.New("destination branch not found")
	}

	resolved := make([]*entity.Product, 0, len(req.ProductIDs))
	for _, pid := range req.ProductIDs {
		productID, err := primitive.ObjectIDFromHex(pid)
		if err != nil {
			return nil, fmt.Errorf("invalid product id: %s", pid)
		}

		product, err := s.productRepo.GetByID(ctx, productID)
		if err != nil || product == nil {
			return nil, fmt.Errorf("product not found: %s", pid)
		}
		if product.BranchID != fromBranchOID {
			return nil, fmt.Errorf("product %s does not belong to source branch", product.Name)
		}
		if product.Status != entity.ProductStatusAvailable {
			return nil, fmt.Errorf("product %s is not available (status: %s)", product.Name, product.Status)
		}

		resolved = append(resolved, product)
	}

	transferNumber, err := s.transferRepo.GenerateTransferNumber(ctx)
	if err != nil {
		return nil, err
	}

	transfer := entity.NewInventoryTransfer(transferNumber, fromBranchOID, toBranchOID, requestedByOID)
	transfer.Notes = req.Notes
	for _, product := range resolved {
		transfer.AddItem(product.ID, 1)
	}

	txErr := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.transferRepo.Create(txCtx, transfer); err != nil {
			return err
		}
		for _, product := range resolved {
			product.Status = entity.ProductStatusReserved
			product.UpdatedAt = time.Now()
			if err := s.productRepo.Update(txCtx, product); err != nil {
				return fmt.Errorf("failed to reserve product %s: %w", product.Name, err)
			}
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return transfer, nil
}

func (s *Service) GetTransfers(ctx context.Context, branchID string, role string) ([]*entity.InventoryTransfer, error) {
	branchOID, err := primitive.ObjectIDFromHex(branchID)
	if err != nil {
		return nil, errors.New("invalid branch ID")
	}

	fromTransfers, _ := s.transferRepo.GetByFromBranchID(ctx, branchOID, nil)
	toTransfers, _ := s.transferRepo.GetByToBranchID(ctx, branchOID, nil)

	return append(fromTransfers, toTransfers...), nil
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

	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		for _, item := range transfer.Items {
			product, err := s.productRepo.GetByID(txCtx, item.ProductID)
			if err != nil || product == nil {
				return fmt.Errorf("product not found during receive: %s", item.ProductID.Hex())
			}
			product.BranchID = transfer.ToBranchID
			product.Status = entity.ProductStatusAvailable
			product.UpdatedAt = time.Now()
			if err := s.productRepo.Update(txCtx, product); err != nil {
				return err
			}
		}
		return s.transferRepo.Update(txCtx, transfer)
	})
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

	if transfer.Status != entity.TransferStatusPending && transfer.Status != entity.TransferStatusInTransit {
		return errors.New("can only cancel pending or in-transit transfers")
	}

	transfer.Cancel()

	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		for _, item := range transfer.Items {
			product, err := s.productRepo.GetByID(txCtx, item.ProductID)
			if err != nil || product == nil {
				return fmt.Errorf("product not found during cancel: %s", item.ProductID.Hex())
			}
			// Restore at the source branch — products were reserved there but never moved.
			product.BranchID = transfer.FromBranchID
			product.Status = entity.ProductStatusAvailable
			product.UpdatedAt = time.Now()
			if err := s.productRepo.Update(txCtx, product); err != nil {
				return err
			}
		}
		return s.transferRepo.Update(txCtx, transfer)
	})
}
