package authorization

import (
	"context"

	"github.com/google/uuid"
)

type AuthorizationService interface {
	CanUpdate(ctx context.Context, userID, videoOwnerID uuid.UUID) error
	CanDelete(ctx context.Context, userID, videoOwnerID uuid.UUID) error
	CanView(ctx context.Context, userID, videoOwnerID uuid.UUID) error
}

type AuthorizationRule interface {
	Check(ctx context.Context, userID, videoOwnerID uuid.UUID) error
}
