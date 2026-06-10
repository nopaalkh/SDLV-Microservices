package usecase

import (
	"context"
	"errors"
	"time"

	"order-service/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance for withdrawal")
	ErrWithdrawalNotFound  = errors.New("withdrawal request not found")
)

type WithdrawalUsecase interface {
	RequestWithdrawal(ctx context.Context, creatorID string, amount float64, bankCode string, accountNum string) (*repository.Withdrawal, error)
	GetCreatorWithdrawals(ctx context.Context, creatorID string) ([]*repository.Withdrawal, error)
	GetAllWithdrawals(ctx context.Context) ([]*repository.Withdrawal, error)
	ApproveWithdrawal(ctx context.Context, withdrawalID string) error
	RejectWithdrawal(ctx context.Context, withdrawalID string) error
}

type withdrawalUsecase struct {
	withdrawalRepo repository.WithdrawalRepository
	orderUsecase   OrderUsecase
}

func NewWithdrawalUsecase(withdrawalRepo repository.WithdrawalRepository, orderUsecase OrderUsecase) WithdrawalUsecase {
	return &withdrawalUsecase{
		withdrawalRepo: withdrawalRepo,
		orderUsecase:   orderUsecase,
	}
}

func (uc *withdrawalUsecase) RequestWithdrawal(ctx context.Context, creatorID string, amount float64, bankCode string, accountNum string) (*repository.Withdrawal, error) {
	if amount <= 0 || bankCode == "" || accountNum == "" {
		return nil, errors.New("invalid withdrawal request data")
	}

	// Calculate total revenue
	totalRevenue, err := uc.orderUsecase.GetCreatorRevenue(ctx, creatorID)
	if err != nil {
		return nil, err
	}

	// Calculate total withdrawn (approved + pending)
	withdrawals, err := uc.withdrawalRepo.GetWithdrawalsByCreator(ctx, creatorID)
	if err != nil {
		return nil, err
	}

	var totalWithdrawn float64
	for _, w := range withdrawals {
		if w.Status == repository.WithdrawalStatusApproved || w.Status == repository.WithdrawalStatusPending {
			totalWithdrawn += w.Amount
		}
	}

	availableBalance := totalRevenue - totalWithdrawn
	if amount > availableBalance {
		return nil, ErrInsufficientBalance
	}

	w := &repository.Withdrawal{
		ID:            uuid.New().String(),
		CreatorID:     creatorID,
		Amount:        amount,
		Status:        repository.WithdrawalStatusPending,
		BankCode:      bankCode,
		AccountNumber: accountNum,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := uc.withdrawalRepo.CreateWithdrawal(ctx, w); err != nil {
		return nil, err
	}

	return w, nil
}

func (uc *withdrawalUsecase) GetCreatorWithdrawals(ctx context.Context, creatorID string) ([]*repository.Withdrawal, error) {
	return uc.withdrawalRepo.GetWithdrawalsByCreator(ctx, creatorID)
}

func (uc *withdrawalUsecase) GetAllWithdrawals(ctx context.Context) ([]*repository.Withdrawal, error) {
	return uc.withdrawalRepo.GetAllWithdrawals(ctx)
}

func (uc *withdrawalUsecase) ApproveWithdrawal(ctx context.Context, withdrawalID string) error {
	w, err := uc.withdrawalRepo.FindWithdrawalByID(ctx, withdrawalID)
	if err != nil {
		return err
	}
	if w == nil {
		return ErrWithdrawalNotFound
	}
	if w.Status != repository.WithdrawalStatusPending {
		return errors.New("only pending withdrawals can be approved")
	}

	// In a real system, you would call Midtrans Payouts / Xendit here.
	return uc.withdrawalRepo.UpdateWithdrawalStatus(ctx, withdrawalID, repository.WithdrawalStatusApproved)
}

func (uc *withdrawalUsecase) RejectWithdrawal(ctx context.Context, withdrawalID string) error {
	w, err := uc.withdrawalRepo.FindWithdrawalByID(ctx, withdrawalID)
	if err != nil {
		return err
	}
	if w == nil {
		return ErrWithdrawalNotFound
	}
	if w.Status != repository.WithdrawalStatusPending {
		return errors.New("only pending withdrawals can be rejected")
	}

	return uc.withdrawalRepo.UpdateWithdrawalStatus(ctx, withdrawalID, repository.WithdrawalStatusRejected)
}
