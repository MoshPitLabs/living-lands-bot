package handlers

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"living-lands-bot/internal/services"
)

type VerifyRequest struct {
	Code           string `json:"code" validate:"required,len=8,alphanum,uppercase"`
	HytaleUsername string `json:"hytale_username" validate:"required,min=3,max=32,alphanum"`
	HytaleUUID     string `json:"hytale_uuid" validate:"required,uuid4"`
}

type VerifyHandler struct {
	account   *services.AccountService
	logger    *slog.Logger
	validator *validator.Validate
}

func NewVerifyHandler(account *services.AccountService, logger *slog.Logger) *VerifyHandler {
	return &VerifyHandler{
		account:   account,
		logger:    logger,
		validator: validator.New(),
	}
}

func (h *VerifyHandler) Handle(c *fiber.Ctx) error {
	var req VerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate input with structured validation
	if err := h.validator.Struct(&req); err != nil {
		validationErrs := err.(validator.ValidationErrors)
		errors := make([]string, len(validationErrs))
		for i, e := range validationErrs {
			errors[i] = formatValidationError(e)
		}

		h.logger.Warn("validation failed",
			"ip", c.IP(),
			"errors", errors,
		)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": errors,
		})
	}

	if err := h.account.VerifyLink(req.Code, req.HytaleUsername, req.HytaleUUID); err != nil {
		h.logger.Error("verify failed",
			"error", err,
			"code", req.Code[:2]+"***", // Partial code for logging
			"username", req.HytaleUsername,
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
	})
}

// formatValidationError converts a validation error to a user-friendly message
func formatValidationError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return err.Field() + " is required"
	case "len":
		return err.Field() + " must be exactly " + err.Param() + " characters"
	case "min":
		return err.Field() + " must be at least " + err.Param() + " characters"
	case "max":
		return err.Field() + " must be at most " + err.Param() + " characters"
	case "alphanum":
		return err.Field() + " must be alphanumeric"
	case "uppercase":
		return err.Field() + " must be uppercase"
	case "uuid4":
		return err.Field() + " must be a valid UUID v4"
	default:
		return err.Field() + " is invalid"
	}
}
