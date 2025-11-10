package shortener

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	shortCodeLength = 5
	alphabet        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Handler struct {
	repo    *Repository
	baseURL string
}

func NewHandler(repo *Repository, baseURL string) *Handler {
	return &Handler{
		repo:    repo,
		baseURL: baseURL,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.POST("/shorten", h.createShortURL)
	}
	r.GET("/:code", h.redirect)
}

type createShortURLRequest struct {
	URL              string `json:"url" binding:"required,url"`
	ExpiresInSeconds *int64 `json:"expires_in_seconds,omitempty"`
}

type createShortURLResponse struct {
	ShortURL    string     `json:"short_url"`
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

func (h *Handler) createShortURL(c *gin.Context) {
	var req createShortURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request: " + err.Error(),
		})
		return
	}
	var expiresAt *time.Time
	if req.ExpiresInSeconds != nil && *req.ExpiresInSeconds > 0 {
		t := time.Now().Add(time.Duration(*req.ExpiresInSeconds) * time.Second)
		expiresAt = &t
	}
	const maxAttempts = 5
	var (
		code string
		err  error
		u    URL
	)
	for i := 0; i < maxAttempts; i++ {
		code, err = generateShortCode()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate short code"})
			return
		}
		u = URL{
			ShortCode:   code,
			OriginalURL: req.URL,
			ExpiresAt:   expiresAt,
		}
		if err = h.repo.Create(c.Request.Context(), &u); err == nil {
			break
		}
		if IsUniqueViolation(err) {
			continue
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error: " + err.Error()})
		return
	}
	if err != nil {
		// repeated collisions – extremely unlikely with 62^5 space
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not allocate short code"})
		return
	}
	shortURL := h.baseURL + "/" + u.ShortCode
	c.JSON(http.StatusCreated, createShortURLResponse{
		ShortURL:    shortURL,
		ShortCode:   u.ShortCode,
		OriginalURL: u.OriginalURL,
		ExpiresAt:   u.ExpiresAt,
	})
}

func (h *Handler) redirect(c *gin.Context) {
	code := c.Param("code")
	if len(code) != shortCodeLength {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	ctx := c.Request.Context()
	u, err := h.repo.GetByShortCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error: " + err.Error()})
		return
	}
	if u.ExpiresAt != nil && time.Now().After(*u.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{
			"error":       "link expired",
			"originalUrl": u.OriginalURL,
		})
		return
	}
	// best-effort click count update (don't block redirect on failure)
	go func(id uint) {
		_ = h.repo.IncrementClick(context.Background(), id)
	}(u.ID)
	c.Redirect(http.StatusTemporaryRedirect, u.OriginalURL)
}

func generateShortCode() (string, error) {
	b := make([]byte, shortCodeLength*2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	code := make([]byte, shortCodeLength)
	for i := 0; i < shortCodeLength; i++ {
		off := i * 2
		val := binary.BigEndian.Uint16(b[off : off+2])
		code[i] = alphabet[int(val)%len(alphabet)]
	}
	return string(code), nil
}
