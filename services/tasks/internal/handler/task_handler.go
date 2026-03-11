package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"tasks-service/internal/service"
	"tech-ip-sem2/shared/models"
)

type TaskHandler struct {
	service *service.TaskService
}

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

func NewTaskHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", http.ErrNoCookie
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", http.ErrNoCookie
	}

	return parts[1], nil
}

// Вспомогательная функция для обработки ошибок аутентификации
func (h *TaskHandler) handleAuthError(w http.ResponseWriter, err error) {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "token invalid"):
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "unauthorized - invalid token"})
	case strings.Contains(errMsg, "auth service timeout"):
		w.WriteHeader(http.StatusGatewayTimeout)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "authentication service timeout"})
	case strings.Contains(errMsg, "auth service unavailable"):
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "authentication service unavailable"})
	case strings.Contains(errMsg, "auth failed"):
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "unauthorized"})
	default:
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal server error"})
	}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	token, err := h.extractToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid authorization"})
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid request format"})
		return
	}

	task, err := h.service.Create(r.Context(), token, req.Title, req.Description, req.DueDate)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	token, err := h.extractToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid authorization"})
		return
	}

	tasks, err := h.service.GetAll(r.Context(), token)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	token, err := h.extractToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid authorization"})
		return
	}

	id := r.PathValue("id")
	task, err := h.service.GetByID(r.Context(), token, id)
	if err != nil {
		if strings.Contains(err.Error(), "auth failed") {
			h.handleAuthError(w, err)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "task not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal server error"})
		return
	}

	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	token, err := h.extractToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid authorization"})
		return
	}

	id := r.PathValue("id")

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid request format"})
		return
	}

	task, err := h.service.Update(r.Context(), token, id, req.Title, req.Done)
	if err != nil {
		if strings.Contains(err.Error(), "auth failed") {
			h.handleAuthError(w, err)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "task not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal server error"})
		return
	}

	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	token, err := h.extractToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "invalid authorization"})
		return
	}

	id := r.PathValue("id")
	err = h.service.Delete(r.Context(), token, id)
	if err != nil {
		if strings.Contains(err.Error(), "auth failed") {
			h.handleAuthError(w, err)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "task not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal server error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
