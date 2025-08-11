package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/jwt"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Interfaces for saving, listing, and getting secrets
type SecretWriter interface {
	Save(ctx context.Context, secretOwner, secretName, secretType string, ciphertext, aesKeyEnc []byte) error
}

type SecretReader interface {
	List(ctx context.Context, secretOwner string) ([]*models.SecretDB, error)

	// Get fetches a single secret by owner, name, and type
	Get(ctx context.Context, secretOwner, secretName, secretType string) (*models.SecretDB, error)
}

type UsernameGetter interface {
	GetUsername(tokenStr string) (string, error)
}

// SecretSaveRequest represents the HTTP request body for saving a secret.
// swagger:model SecretSaveRequest
type SecretSaveRequest struct {
	// Secret name
	// required: true
	Name string `json:"name"`
	// Secret type
	// required: true
	Type string `json:"type"`
	// Encrypted secret data
	// required: true
	Ciphertext []byte `json:"ciphertext"`
	// Encrypted AES key
	// required: true
	AESKeyEnc []byte `json:"aes_key_enc"`
}

// SecretResponse represents the secret data returned in responses.
// swagger:model SecretResponse
type SecretResponse struct {
	SecretName  string    `json:"secret_name"`
	SecretType  string    `json:"secret_type"`
	SecretOwner string    `json:"secret_owner"`
	Ciphertext  []byte    `json:"ciphertext"`
	AESKeyEnc   []byte    `json:"aes_key_enc"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SecretListResponse represents a list of secrets in responses.
// swagger:model SecretListResponse
type SecretListResponse struct {
	Secrets []SecretResponse `json:"secrets"`
}

// --- HTTP Handlers ---

type SecretWriteHTTPHandler struct {
	writer SecretWriter
	jwt    UsernameGetter
}

func NewSecretWriteHTTPHandler(writer SecretWriter, jwtHandler UsernameGetter) *SecretWriteHTTPHandler {
	return &SecretWriteHTTPHandler{writer: writer, jwt: jwtHandler}
}

// Save secret handler
// @Summary Save a secret
// @Description Saves a secret for authenticated user
// @Tags secrets
// @Accept json
// @Produce json
// @Param secret body SecretSaveRequest true "Secret save request payload"
// @Success 200 {string} string "ok"
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "internal server error"
// @Router /secrets [post]
func (h *SecretWriteHTTPHandler) Save(w http.ResponseWriter, r *http.Request) {
	token, err := jwt.GetTokenFromHeader(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	username, err := h.jwt.GetUsername(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req SecretSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.writer.Save(r.Context(), username, req.Name, req.Type, req.Ciphertext, req.AESKeyEnc); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type SecretReadHTTPHandler struct {
	reader SecretReader
	jwt    UsernameGetter
}

func NewSecretReadHTTPHandler(reader SecretReader, jwtHandler UsernameGetter) *SecretReadHTTPHandler {
	return &SecretReadHTTPHandler{reader: reader, jwt: jwtHandler}
}

// List secrets handler
// @Summary List all secrets
// @Description Lists all secrets for authenticated user
// @Tags secrets
// @Accept json
// @Produce json
// @Success 200 {object} SecretListResponse
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "internal server error"
// @Router /secrets [get]
func (h *SecretReadHTTPHandler) List(w http.ResponseWriter, r *http.Request) {
	token, err := jwt.GetTokenFromHeader(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	username, err := h.jwt.GetUsername(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	secrets, err := h.reader.List(r.Context(), username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := SecretListResponse{
		Secrets: make([]SecretResponse, 0, len(secrets)),
	}

	for _, s := range secrets {
		resp.Secrets = append(resp.Secrets, SecretResponse{
			SecretName:  s.SecretName,
			SecretType:  s.SecretType,
			SecretOwner: s.SecretOwner,
			Ciphertext:  s.Ciphertext,
			AESKeyEnc:   s.AESKeyEnc,
			CreatedAt:   s.CreatedAt,
			UpdatedAt:   s.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Get single secret handler
// @Summary Get a secret
// @Description Gets a secret by name and type for authenticated user
// @Tags secrets
// @Accept json
// @Produce json
// @Param name query string true "Secret name"
// @Param type query string true "Secret type"
// @Success 200 {object} SecretResponse
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "unauthorized"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "internal server error"
// @Router /secrets/get [get]
func (h *SecretReadHTTPHandler) Get(w http.ResponseWriter, r *http.Request) {
	token, err := jwt.GetTokenFromHeader(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	username, err := h.jwt.GetUsername(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	secretName := r.URL.Query().Get("name")
	secretType := r.URL.Query().Get("type")

	if secretName == "" || secretType == "" {
		http.Error(w, "missing name or type query parameter", http.StatusBadRequest)
		return
	}

	secret, err := h.reader.Get(r.Context(), username, secretName, secretType)
	if err != nil {
		http.Error(w, "secret not found", http.StatusNotFound)
		return
	}

	resp := SecretResponse{
		SecretName:  secret.SecretName,
		SecretType:  secret.SecretType,
		SecretOwner: secret.SecretOwner,
		Ciphertext:  secret.Ciphertext,
		AESKeyEnc:   secret.AESKeyEnc,
		CreatedAt:   secret.CreatedAt,
		UpdatedAt:   secret.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type SecretWriteGRPCHandler struct {
	pb.UnimplementedSecretWriteServiceServer
	writer SecretWriter
	jwt    UsernameGetter
}

func NewSecretWriteGRPCHandler(
	writer SecretWriter,
	jwt UsernameGetter,
) *SecretWriteGRPCHandler {
	return &SecretWriteGRPCHandler{
		writer: writer,
		jwt:    jwt,
	}
}

func (h *SecretWriteGRPCHandler) Save(ctx context.Context, req *pb.SecretSaveRequest) (*emptypb.Empty, error) {
	token, err := jwt.GetTokenFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid auth token")
	}

	username, err := h.jwt.GetUsername(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	if err := h.writer.Save(ctx, username, req.SecretName, req.SecretType, req.Ciphertext, req.AesKeyEnc); err != nil {
		return nil, status.Error(codes.Internal, "failed to save secret")
	}

	return &emptypb.Empty{}, nil
}

type SecretReadGRPCHandler struct {
	pb.UnimplementedSecretReadServiceServer
	reader SecretReader
	jwt    UsernameGetter
}

func NewSecretReadGRPCHandler(
	reader SecretReader,
	jwt UsernameGetter,
) *SecretReadGRPCHandler {
	return &SecretReadGRPCHandler{
		reader: reader,
		jwt:    jwt,
	}
}

func (h *SecretReadGRPCHandler) Get(ctx context.Context, req *pb.SecretGetRequest) (*pb.Secret, error) {
	token, err := jwt.GetTokenFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid auth token")
	}

	username, err := h.jwt.GetUsername(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	secrets, err := h.reader.List(ctx, username)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list secrets")
	}

	for _, s := range secrets {
		if s.SecretName == req.SecretName && s.SecretType == req.SecretType {
			return &pb.Secret{
				SecretName:  s.SecretName,
				SecretType:  s.SecretType,
				SecretOwner: s.SecretOwner,
				Ciphertext:  s.Ciphertext,
				AesKeyEnc:   s.AESKeyEnc,
				CreatedAt:   timestamppb.New(s.CreatedAt),
				UpdatedAt:   timestamppb.New(s.UpdatedAt),
			}, nil
		}
	}

	return nil, status.Error(codes.NotFound, "secret not found")
}

func (h *SecretReadGRPCHandler) List(_ *emptypb.Empty, stream pb.SecretReadService_ListServer) error {
	ctx := stream.Context()

	token, err := jwt.GetTokenFromContext(ctx)
	if err != nil {
		return status.Error(codes.Unauthenticated, "missing or invalid auth token")
	}

	username, err := h.jwt.GetUsername(token)
	if err != nil {
		return status.Error(codes.Unauthenticated, "invalid token")
	}

	secrets, err := h.reader.List(ctx, username)
	if err != nil {
		return status.Error(codes.Internal, "failed to list secrets")
	}

	for _, s := range secrets {
		resp := &pb.Secret{
			SecretName:  s.SecretName,
			SecretType:  s.SecretType,
			SecretOwner: s.SecretOwner,
			Ciphertext:  s.Ciphertext,
			AesKeyEnc:   s.AESKeyEnc,
			CreatedAt:   timestamppb.New(s.CreatedAt),
			UpdatedAt:   timestamppb.New(s.UpdatedAt),
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
	}

	return nil
}
