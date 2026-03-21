package auth

import (
	"context"
	"errors"
)

type contextKey string

const (
	userIDKey   contextKey = "userID"
	userRoleKey contextKey = "userRole"
)

// -- GETTERS (for the handlers)

func GetUserID(ctx context.Context) (string, error) {
	val := ctx.Value(userIDKey)

	// We use the Type Assertion .(string) again because Context values are stored as empty interfaces (any)
	id, ok := val.(string) // Type Assertion safely convert from 'any' to 'string'
	if !ok {
		return "", errors.New("user ID not found in context")
	}
	return id, nil
}

func GetUserRole(ctx context.Context) (string, error) {
	val := ctx.Value(userRoleKey)
	role, ok := val.(string)
	if !ok {
		return "", errors.New("user Role not found in context")
	}
	return role, nil
}

// -- SETTERS (for the middleware)

func ContextWithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func ContextWithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, userRoleKey, role)
}
