package authorization

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type OwnershipRule struct{}

func NewOwnershipRule() *OwnershipRule {
	return &OwnershipRule{}
}

func (r *OwnershipRule) Check(ctx context.Context, userID, videoOwnerID uuid.UUID) error {
	if userID != videoOwnerID {
		return fmt.Errorf("unauthorized: only owner can perform this action")
	}
	return nil
}

type PublicAccessRule struct{}

func NewPublicAccessRule() *PublicAccessRule {
	return &PublicAccessRule{}
}

func (r *PublicAccessRule) Check(ctx context.Context, userID, videoOwnerID uuid.UUID) error {
	return nil
}

type AdminRule struct {
	adminIDs map[uuid.UUID]bool
}

func NewAdminRule(adminIDs []uuid.UUID) *AdminRule {
	admins := make(map[uuid.UUID]bool)
	for _, id := range adminIDs {
		admins[id] = true
	}
	return &AdminRule{adminIDs: admins}
}

func (r *AdminRule) Check(ctx context.Context, userID, videoOwnerID uuid.UUID) error {
	if r.adminIDs[userID] {
		return nil
	}
	return fmt.Errorf("unauthorized: admin access required")
}

type CompositeRule struct {
	rules []AuthorizationRule
}

func NewCompositeRule(rules ...AuthorizationRule) *CompositeRule {
	return &CompositeRule{rules: rules}
}

func (r *CompositeRule) Check(ctx context.Context, userID, videoOwnerID uuid.UUID) error {
	var lastErr error
	for _, rule := range r.rules {
		if err := rule.Check(ctx, userID, videoOwnerID); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	return lastErr
}
