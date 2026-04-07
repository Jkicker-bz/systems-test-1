package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/Jkicker-bz/systems_test_1/internal/helpers"
	"github.com/Jkicker-bz/systems_test_1/internal/models"
)

// GET /set_lists
// ListSetLists fetches every set list ordered by date descending.
// COALESCE converts any NULL column into an empty string so the
// plain string fields in models.SetList are always safe to scan into.
func (app *Application) ListSetLists(w http.ResponseWriter, r *http.Request) {
	query := `
SELECT id, name,
COALESCE(event_type, ''),
COALESCE(date::text, ''),
COALESCE(notes, ''),
COALESCE(created_by::text, ''),
created_at
FROM set_lists
ORDER BY date DESC`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := app.DB.QueryContext(ctx, query)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	defer rows.Close()

	var setLists []models.SetList

	for rows.Next() {
		var sl models.SetList
		err := rows.Scan(
			&sl.ID, &sl.Name, &sl.EventType, &sl.Date,
			&sl.Notes, &sl.CreatedBy, &sl.CreatedAt,
		)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		setLists = append(setLists, sl)
	}

	if err = rows.Err(); err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"set_lists": setLists}, nil)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// GET /set_lists/{id}
// GetSetList fetches a single set list by primary key.
func (app *Application) GetSetList(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		helpers.NotFound(w)
		return
	}

	query := `
SELECT id, name,
COALESCE(event_type, ''),
COALESCE(date::text, ''),
COALESCE(notes, ''),
COALESCE(created_by::text, ''),
created_at
FROM set_lists
WHERE id = $1`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var sl models.SetList
	err := app.DB.QueryRowContext(ctx, query, id).Scan(
		&sl.ID, &sl.Name, &sl.EventType, &sl.Date,
		&sl.Notes, &sl.CreatedBy, &sl.CreatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			helpers.NotFound(w)
		default:
			helpers.ServerError(w, err)
		}
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"set_list": sl}, nil)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// POST /set_lists
// CreateSetList inserts a new set list and returns the generated row.
func (app *Application) CreateSetList(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name      string `json:"name"`
		EventType string `json:"event_type"`
		Date      string `json:"date"`
		Notes     string `json:"notes"`
		CreatedBy string `json:"created_by"`
	}

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		helpers.BadRequest(w, err.Error())
		return
	}

	v := helpers.NewValidator()
	v.Check(input.Name != "", "name", "must be provided")

	if !v.Valid() {
		helpers.FailedValidation(w, v.Errors)
		return
	}

	query := `
INSERT INTO set_lists (name, event_type, date, notes, created_by)
VALUES ($1, $2, NULLIF($3, '')::date, $4, NULLIF($5, '')::uuid)
RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var sl models.SetList
	sl.Name = input.Name
	sl.EventType = input.EventType
	sl.Date = input.Date
	sl.Notes = input.Notes
	sl.CreatedBy = input.CreatedBy

	err = app.DB.QueryRowContext(ctx, query,
		input.Name, input.EventType, input.Date, input.Notes, input.CreatedBy,
	).Scan(&sl.ID, &sl.CreatedAt)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	extra := http.Header{"Location": []string{"/set_lists/" + sl.ID}}
	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"set_list": sl}, extra)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// PUT /set_lists/{id}
// UpdateSetList replaces all fields on an existing set list row.
func (app *Application) UpdateSetList(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		helpers.NotFound(w)
		return
	}

	var input struct {
		Name      string `json:"name"`
		EventType string `json:"event_type"`
		Date      string `json:"date"`
		Notes     string `json:"notes"`
		CreatedBy string `json:"created_by"`
	}

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		helpers.BadRequest(w, err.Error())
		return
	}

	v := helpers.NewValidator()
	v.Check(input.Name != "", "name", "must be provided")

	if !v.Valid() {
		helpers.FailedValidation(w, v.Errors)
		return
	}

	query := `
UPDATE set_lists
SET name = $1, event_type = $2, date = NULLIF($3, '')::date,
   notes = $4, created_by = NULLIF($5, '')::uuid
WHERE id = $6`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result, err := app.DB.ExecContext(ctx, query,
		input.Name, input.EventType, input.Date, input.Notes, input.CreatedBy, id,
	)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	if rowsAffected == 0 {
		helpers.NotFound(w)
		return
	}

	updated := models.SetList{
		ID: id, Name: input.Name, EventType: input.EventType,
		Date: input.Date, Notes: input.Notes, CreatedBy: input.CreatedBy,
	}
	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"set_list": updated}, nil)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// DELETE /set_lists/{id}
// DeleteSetList removes a set list row.
// Returns 204 No Content on success.
func (app *Application) DeleteSetList(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		helpers.NotFound(w)
		return
	}

	query := `DELETE FROM set_lists WHERE id = $1`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result, err := app.DB.ExecContext(ctx, query, id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	if rowsAffected == 0 {
		helpers.NotFound(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
