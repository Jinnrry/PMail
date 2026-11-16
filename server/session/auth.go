package session

import (
	"context"

	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/models"
)

// Login rotates the session token and stores only the identity and credential
// version. Permissions are read from the database on each authenticated request.
func Login(ctx context.Context, user *models.User) error {
	if err := Instance.RenewToken(ctx); err != nil {
		return err
	}
	if err := Instance.Clear(ctx); err != nil {
		return err
	}
	Instance.Put(ctx, "user_id", user.ID)
	Instance.Put(ctx, "auth_version", user.AuthVersion)
	return nil
}

// CurrentUser rejects revoked sessions, including legacy sessions that did not
// record a credential version. A nil user means the request is unauthenticated.
func CurrentUser(ctx context.Context) (*models.User, error) {
	userID := Instance.GetInt(ctx, "user_id")
	if userID <= 0 || !Instance.Exists(ctx, "auth_version") {
		if userID > 0 || Instance.Exists(ctx, "user") {
			return nil, Instance.Destroy(ctx)
		}
		return nil, nil
	}
	var user models.User
	found, err := db.Instance.ID(userID).Get(&user)
	if err != nil {
		return nil, err
	}
	if !found || user.Disabled != 0 || user.AuthVersion != Instance.GetInt64(ctx, "auth_version") {
		return nil, Instance.Destroy(ctx)
	}
	return &user, nil
}
