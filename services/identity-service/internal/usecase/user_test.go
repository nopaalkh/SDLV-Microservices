package usecase

import (
	"context"
	"testing"
	"time"

	"identity-service/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *repository.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindUserByEmail(ctx context.Context, email string) (*repository.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockUserRepository) FindUserByID(ctx context.Context, id string) (*repository.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *repository.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetAllUsers(ctx context.Context) ([]*repository.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUserStatus(ctx context.Context, id string, status repository.UserStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func TestRegister(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := NewAuthUsecase(mockRepo, "secret", time.Hour)

	// Test Case 1: Successful Registration
	t.Run("Success", func(t *testing.T) {
		mockRepo.On("FindUserByEmail", mock.Anything, "newuser@test.com").Return(nil, nil).Once()
		mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*repository.User")).Return(nil).Once()

		user, err := uc.Register(context.Background(), "newuser@test.com", "password123", repository.RoleBuyer)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "newuser@test.com", user.Email)
		assert.Equal(t, repository.RoleBuyer, user.Role)
		mockRepo.AssertExpectations(t)
	})

	// Test Case 2: Email Already Exists
	t.Run("EmailAlreadyExists", func(t *testing.T) {
		existingUser := &repository.User{ID: "1", Email: "existing@test.com"}
		mockRepo.On("FindUserByEmail", mock.Anything, "existing@test.com").Return(existingUser, nil).Once()

		user, err := uc.Register(context.Background(), "existing@test.com", "password123", repository.RoleBuyer)

		assert.ErrorIs(t, err, ErrUserAlreadyExists)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetUserByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := NewAuthUsecase(mockRepo, "secret", time.Hour)

	// Test Case 1: User Found
	t.Run("Found", func(t *testing.T) {
		expectedUser := &repository.User{ID: "123", Email: "user@test.com"}
		mockRepo.On("FindUserByID", mock.Anything, "123").Return(expectedUser, nil).Once()

		user, err := uc.GetUserByID(context.Background(), "123")

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})

	// Test Case 2: User Not Found
	t.Run("NotFound", func(t *testing.T) {
		mockRepo.On("FindUserByID", mock.Anything, "999").Return(nil, nil).Once()

		user, err := uc.GetUserByID(context.Background(), "999")

		assert.ErrorIs(t, err, ErrUserNotFound)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}
