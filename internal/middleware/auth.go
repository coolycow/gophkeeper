// Package middleware содержит HTTP-middleware: аутентификация, gzip, логирование ошибок.
package middleware

import (
	"context"
	"errors"
	"net/http"

	httpError "github.com/coolycow/gophkeeper/internal/error"
	"github.com/coolycow/gophkeeper/internal/model"
	"github.com/coolycow/gophkeeper/internal/service"
	"github.com/gin-gonic/gin"
)

type ginKey string

const (
	UserIDKey ginKey = "userID"
)

// GetUserIDFromGinContext получает userID из Gin контекста
func GetUserIDFromGinContext(c *gin.Context) (string, error) {
	value, exists := c.Get(string(UserIDKey))

	if !exists || value == nil {
		return "", errors.New("user ID not found in context")
	}

	userID, ok := value.(string)

	if !ok {
		return "", errors.New("incorrect user ID in context")
	}

	return userID, nil
}

// createUserAndCookieValue создает пользователя и получает значение куки для него
func createUserAndCookieValue(ctx context.Context, userService service.UserService) (model.User, string, error) {
	user, err := userService.CreateUser(ctx)
	if err != nil {
		return model.User{}, "", err
	}

	cookieValue, err := userService.GetCookieValueByUser(*user)

	if err != nil {
		return model.User{}, "", err
	}

	return *user, cookieValue, nil
}

// OptionalAuthMiddleware проверяет наличие куки авторизации и создает пользователя если куки нет
func OptionalAuthMiddleware(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("auth")

		if err != nil {
			user, cookieValue, err := createUserAndCookieValue(context.Background(), userService)

			if err != nil {
				_ = c.Error(httpError.CustomError{
					Message:    err.Error(),
					StatusCode: http.StatusInternalServerError,
				})
				c.Abort()
				return
			}

			http.SetCookie(c.Writer, &http.Cookie{
				Name:     "auth",
				Value:    cookieValue,
				Path:     "/",
				HttpOnly: true,
			})

			c.Set(string(UserIDKey), user.ID)
		} else {
			// Проверяем валидность куки
			userID, err := userService.GetUserIDFromCookie(cookie)

			if err != nil {
				user, cookieValue, createErr := createUserAndCookieValue(context.Background(), userService)

				if createErr != nil {
					_ = c.Error(httpError.CustomError{
						Message:    createErr.Error(),
						StatusCode: http.StatusInternalServerError,
					})
					c.Abort()
					return
				}

				http.SetCookie(c.Writer, &http.Cookie{
					Name:     "auth",
					Value:    cookieValue,
					Path:     "/",
					HttpOnly: true,
				})

				userID = user.ID
			}

			_, err = userService.GetUserByID(c.Request.Context(), userID)
			if err != nil {
				_ = c.Error(httpError.CustomError{
					Message:    "User with this ID does not exist",
					StatusCode: http.StatusUnauthorized,
				})
				c.Abort()
				return
			}

			c.Set(string(UserIDKey), userID)
		}

		c.Next()
	}
}

// RequiredAuthMiddleware проверяет наличие куки авторизации и получает userID из куки
func RequiredAuthMiddleware(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("auth")

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		userID, err := userService.GetUserIDFromCookie(cookie)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid cookie"})
			c.Abort()
			return
		}

		_, err = userService.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			_ = c.Error(httpError.CustomError{
				Message:    "User with this ID does not exist",
				StatusCode: http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		c.Set(string(UserIDKey), userID)

		c.Next()
	}
}
