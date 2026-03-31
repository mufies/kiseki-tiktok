package authorization

import (
	"context"

	"github.com/google/uuid"
)

type VideoAuthorizationService struct {
	updateRule AuthorizationRule
	deleteRule AuthorizationRule
	viewRule   AuthorizationRule
}

func NewVideoAuthorizationService() *VideoAuthorizationService {
	return &VideoAuthorizationService{
		updateRule: NewOwnershipRule(),
		deleteRule: NewOwnershipRule(),
		viewRule:   NewPublicAccessRule(),
	}
}

func NewVideoAuthorizationServiceWithRules(
	updateRule, deleteRule, viewRule AuthorizationRule,
) *VideoAuthorizationService {
	return &VideoAuthorizationService{
		updateRule: updateRule,
		deleteRule: deleteRule,
		viewRule:   viewRule,
	}
}

func (s *VideoAuthorizationService) CanUpdate(ctx context.Context, userID, videoOwnerID uuid.UUID) error {
	return s.updateRule.Check(ctx, userID, videoOwnerID)
}

func (s *VideoAuthorizationService) CanDelete(ctx context.Context, userID, videoOwnerID uuid.UUID) error {
	return s.deleteRule.Check(ctx, userID, videoOwnerID)
}

func (s *VideoAuthorizationService) CanView(ctx context.Context, userID, videoOwnerID uuid.UUID) error {
	return s.viewRule.Check(ctx, userID, videoOwnerID)
}
