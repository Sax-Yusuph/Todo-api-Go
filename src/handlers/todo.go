package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sax-yusuph/todo/models"
)

type Todo struct {
	Client *redis.Client
}

func getTodosKey(userID string) string {
	return fmt.Sprintf("user:%s:todos", userID)
}

func (a *Todo) Create(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	userKey := getTodosKey(userID)

	var todo models.Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Generate a new ID and ensure default completed state
	todo.ID = uuid.New().String()
	todo.Completed = false

	// Marshal todo to JSON to store in Redis
	todoJSON, err := json.Marshal(todo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to serialize todo")
		return
	}

	// Store in the user's hash: HSET user:userA:todos <todoID> <todoJSON>
	if err := a.Client.HSet(r.Context(), userKey, todo.ID, todoJSON).Err(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save todo")
		return
	}

	respondWithJSON(w, http.StatusCreated, todo)
}

func (a *Todo) List(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	userKey := getTodosKey(userID)

	// Get all fields (todo JSON strings) from the user's hash: HGETALL user:userA:todos
	todoMap, err := a.Client.HGetAll(r.Context(), userKey).Result()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve todos")
		return
	}

	todos := make([]models.Todo, 0, len(todoMap))
	for _, todoJSON := range todoMap {
		var todo models.Todo
		if err := json.Unmarshal([]byte(todoJSON), &todo); err != nil {
			log.Printf("Failed to unmarshal todo: %v", err)
			continue
		}
		todos = append(todos, todo)
	}

	fmt.Println("reached here")
	respondWithJSON(w, http.StatusOK, todos)
}

func (a *Todo) GetById(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	todoID := chi.URLParam(r, "todoID")
	userKey := getTodosKey(userID)

	todoJSON, err := a.Client.HGet(r.Context(), userKey, todoID).Result()
	if errors.Is(err, redis.Nil) {
		respondWithError(w, http.StatusNotFound, "Todo not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve todo")
		return
	}

	var todo Todo
	if err := json.Unmarshal([]byte(todoJSON), &todo); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to parse todo data")
		return
	}

	respondWithJSON(w, http.StatusOK, todo)
}

func (a *Todo) UpdateById(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	todoID := chi.URLParam(r, "todoID")
	userKey := getTodosKey(userID)

	if err := a.Client.HExists(r.Context(), userKey, todoID).Err(); err != nil {
		respondWithError(w, http.StatusNotFound, "Todo not found")
		return
	}

	var updateData struct {
		Task      string `json:"task"`
		Completed bool   `json:"completed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Create the full Todo object to store
	updatedTodo := models.Todo{
		ID:        todoID, // Keep the original ID
		Task:      updateData.Task,
		Completed: updateData.Completed,
	}

	todoJSON, err := json.Marshal(updatedTodo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to serialize todo")
		return
	}

	// Overwrite the existing field in the hash: HSET user:userA:todos <todoID> <newTodoJSON>
	if err := a.Client.HSet(r.Context(), userKey, todoID, todoJSON).Err(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update todo")
		return
	}

	respondWithJSON(w, http.StatusOK, updatedTodo)
}

func (a *Todo) DeleteById(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	todoID := chi.URLParam(r, "todoID")
	userKey := getTodosKey(userID)

	deletedCount, err := a.Client.HDel(r.Context(), userKey, todoID).Result()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete todo")
		return
	}

	if deletedCount == 0 {
		respondWithError(w, http.StatusNotFound, "Todo not found")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil) // 204 No Content
}
