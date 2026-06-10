package usecase

import (
	"context"
	"errors"
	"time"

	"catalog-service/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrNotPurchased = errors.New("user has not purchased this asset")
)

type ReviewResponse struct {
	ID        string    `json:"id"`
	AssetID   string    `json:"asset_id"`
	UserID    string    `json:"user_id"`
	UserEmail string    `json:"user_email"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type ReviewUsecase interface {
	CreateReview(ctx context.Context, assetID string, userID string, rating int, comment string) (*ReviewResponse, error)
	GetReviewsByAssetID(ctx context.Context, assetID string) ([]*ReviewResponse, error)
}

type reviewUsecase struct {
	reviewRepo     repository.ReviewRepository
	orderClient    OrderClient
	identityClient IdentityClient
}

func NewReviewUsecase(reviewRepo repository.ReviewRepository, orderClient OrderClient, identityClient IdentityClient) ReviewUsecase {
	return &reviewUsecase{
		reviewRepo:     reviewRepo,
		orderClient:    orderClient,
		identityClient: identityClient,
	}
}

func (uc *reviewUsecase) CreateReview(ctx context.Context, assetID string, userID string, rating int, comment string) (*ReviewResponse, error) {
	if rating < 1 || rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}

	// Verify purchase
	hasPurchased, err := uc.orderClient.HasPurchased(userID, assetID)
	if err != nil {
		return nil, err
	}
	if !hasPurchased {
		return nil, ErrNotPurchased
	}

	rev := &repository.Review{
		ID:        uuid.New().String(),
		AssetID:   assetID,
		UserID:    userID,
		Rating:    rating,
		Comment:   comment,
		CreatedAt: time.Now(),
	}

	if err := uc.reviewRepo.CreateReview(ctx, rev); err != nil {
		return nil, err
	}

	email, err := uc.identityClient.GetUserEmail(userID)
	if err != nil || email == "" {
		email = "Pengguna"
	}

	return &ReviewResponse{
		ID:        rev.ID,
		AssetID:   rev.AssetID,
		UserID:    rev.UserID,
		UserEmail: email,
		Rating:    rev.Rating,
		Comment:   rev.Comment,
		CreatedAt: rev.CreatedAt,
	}, nil
}

func (uc *reviewUsecase) GetReviewsByAssetID(ctx context.Context, assetID string) ([]*ReviewResponse, error) {
	reviews, err := uc.reviewRepo.GetReviewsByAssetID(ctx, assetID)
	if err != nil {
		return nil, err
	}

	var response []*ReviewResponse
	for _, r := range reviews {
		email, err := uc.identityClient.GetUserEmail(r.UserID)
		if err != nil || email == "" {
			email = "Pengguna"
		}

		response = append(response, &ReviewResponse{
			ID:        r.ID,
			AssetID:   r.AssetID,
			UserID:    r.UserID,
			UserEmail: email,
			Rating:    r.Rating,
			Comment:   r.Comment,
			CreatedAt: r.CreatedAt,
		})
	}

	return response, nil
}
