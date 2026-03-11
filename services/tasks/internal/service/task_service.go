package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"tasks-service/internal/client"

	"github.com/google/uuid"
)

type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     string    `json:"due_date"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TaskService struct {
	authClient *client.AuthClient
	tasks      map[string]Task
	mu         sync.RWMutex
}

func NewTaskService(authClient *client.AuthClient) *TaskService {
	return &TaskService{
		authClient: authClient,
		tasks:      make(map[string]Task),
	}
}

func (s *TaskService) Create(ctx context.Context, token string, title, description, dueDate string) (Task, error) {
	println("[TaskService] Calling VerifyToken via gRPC")
	username, err := s.authClient.VerifyToken(ctx, token)
	if err != nil {
		return Task{}, fmt.Errorf("auth failed: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	task := Task{
		ID:          "t_" + uuid.New().String()[:8],
		Title:       title,
		Description: description,
		DueDate:     dueDate,
		Done:        false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.tasks[task.ID] = task
	_ = username

	return task, nil
}

func (s *TaskService) GetAll(ctx context.Context, token string) ([]Task, error) {
	println("[TaskService] Calling VerifyToken via gRPC for GetAll")
	_, err := s.authClient.VerifyToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("auth failed: %w", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s *TaskService) GetByID(ctx context.Context, token string, id string) (Task, error) {
	println("[TaskService] Calling VerifyToken via gRPC for GetByID")
	_, err := s.authClient.VerifyToken(ctx, token)
	if err != nil {
		return Task{}, fmt.Errorf("auth failed: %w", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return Task{}, fmt.Errorf("task not found")
	}

	return task, nil
}

func (s *TaskService) Update(ctx context.Context, token string, id string, title *string, done *bool) (Task, error) {
	println("[TaskService] Calling VerifyToken via gRPC for Update")
	_, err := s.authClient.VerifyToken(ctx, token)
	if err != nil {
		return Task{}, fmt.Errorf("auth failed: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return Task{}, fmt.Errorf("task not found")
	}

	if title != nil {
		task.Title = *title
	}
	if done != nil {
		task.Done = *done
	}
	task.UpdatedAt = time.Now()

	s.tasks[id] = task
	return task, nil
}

func (s *TaskService) Delete(ctx context.Context, token string, id string) error {
	println("[TaskService] Calling VerifyToken via gRPC for Delete")
	_, err := s.authClient.VerifyToken(ctx, token)
	if err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return fmt.Errorf("task not found")
	}

	delete(s.tasks, id)
	return nil
}
