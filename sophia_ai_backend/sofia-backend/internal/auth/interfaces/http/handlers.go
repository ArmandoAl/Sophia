package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/armandoalvarado/sofia-backend/internal/auth/application"
	"github.com/armandoalvarado/sofia-backend/internal/auth/domain"
	authjwt "github.com/armandoalvarado/sofia-backend/internal/auth/infrastructure/jwt"
	"github.com/armandoalvarado/sofia-backend/internal/platform/httpjson"
)

type AuthHandler struct {
	registerUC     *application.RegisterUserUseCase
	loginUC        *application.LoginUserUseCase
	tokenService   *authjwt.Service
	requestBodyMax int64
	postRegister   func(*domain.User) error
	audit          LoginFailureAuditRecorder
}

type LoginFailureAuditRecorder interface {
	RecordAuditLog(ctx context.Context, userID, action, resourceType, resourceID string, metadata map[string]any) error
}

func NewAuthHandler(
	registerUC *application.RegisterUserUseCase,
	loginUC *application.LoginUserUseCase,
	tokenService *authjwt.Service,
	requestBodyMax int64,
) *AuthHandler {
	return &AuthHandler{
		registerUC:     registerUC,
		loginUC:        loginUC,
		tokenService:   tokenService,
		requestBodyMax: requestBodyMax,
	}
}

func (h *AuthHandler) SetPostRegisterHook(hook func(*domain.User) error) {
	h.postRegister = hook
}

func (h *AuthHandler) SetAuditRecorder(audit LoginFailureAuditRecorder) {
	h.audit = audit
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.MethodNotAllowed(w)
		return
	}

	var req RegisterRequest
	if err := httpjson.DecodeJSON(w, r, &req, h.requestBodyMax); err != nil {
		httpjson.BadRequest(w, "invalid request body")
		return
	}

	user, err := h.registerUC.Execute(req.Name, req.Email, req.Password)
	if err != nil {
		h.handleError(w, err)
		return
	}
	if h.postRegister != nil {
		if err := h.postRegister(user); err != nil {
			httpjson.InternalServerError(w)
			return
		}
	}

	resp := UserResponse{ID: user.ID, Name: user.Name, Email: user.Email}
	httpjson.WriteJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.MethodNotAllowed(w)
		return
	}

	var req LoginRequest
	if err := httpjson.DecodeJSON(w, r, &req, h.requestBodyMax); err != nil {
		httpjson.BadRequest(w, "invalid request body")
		return
	}

	user, err := h.loginUC.Execute(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			h.recordLoginFailure(r.Context(), req.Email, r)
		}
		h.handleError(w, err)
		return
	}

	token, err := h.tokenService.Generate(user.ID, user.Email, user.Role)
	if err != nil {
		httpjson.InternalServerError(w)
		return
	}

	resp := LoginResponse{
		Token: token,
		User:  UserResponse{ID: user.ID, Name: user.Name, Email: user.Email},
	}
	httpjson.WriteJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) recordLoginFailure(ctx context.Context, email string, r *http.Request) {
	if h.audit == nil {
		return
	}
	_ = h.audit.RecordAuditLog(ctx, "", "auth_login_failed", "auth", "login", map[string]any{
		"email":      strings.ToLower(strings.TrimSpace(email)),
		"remote_ip":  remoteIP(r),
		"user_agent": truncate(r.UserAgent(), 120),
	})
}

func remoteIP(r *http.Request) string {
	forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwarded != "" {
		if idx := strings.Index(forwarded, ","); idx >= 0 {
			forwarded = forwarded[:idx]
		}
		return strings.TrimSpace(forwarded)
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func truncate(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max] + "...[TRUNCATED]"
}

func (h *AuthHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		httpjson.WriteError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		httpjson.Unauthorized(w)
	case errors.Is(err, domain.ErrInvalidName),
		errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrInvalidPassword),
		errors.Is(err, domain.ErrWeakPassword):
		httpjson.BadRequest(w, err.Error())
	default:
		httpjson.InternalServerError(w)
	}
}
