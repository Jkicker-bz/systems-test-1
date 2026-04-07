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

// Application holds the database connection pool.
// All handlers are methods on this struct so they can access the DB.
type Application struct {
	DB *sql.DB
}

// GET /leads
// ListLeads fetches every lead row ordered by name.
// COALESCE converts any NULL column into an empty string so the
// plain string fields in models.Lead are always safe to scan into.
func (app *Application) ListLeads(w http.ResponseWriter, r *http.Request) {
	query := `
SELECT id, name,
COALESCE(role, ''),
COALESCE(color, ''),
created_at
FROM leads
ORDER BY name`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := app.DB.QueryContext(ctx, query)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	defer rows.Close()

	var leads []models.Lead

	for rows.Next() {
		var l models.Lead
		err := rows.Scan(&l.ID, &l.Name, &l.Role, &l.Color, &l.CreatedAt)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		leads = append(leads, l)
	}

	if err = rows.Err(); err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"leads": leads}, nil)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// GET /leads/{id}
// GetLead fetches a single lead by primary key.
func (app *Application) GetLead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		helpers.NotFound(w)
		return
	}

	query := `
SELECT id, name,
COALESCE(role, ''),
COALESCE(color, ''),
created_at
FROM leads
WHERE id = $1`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var l models.Lead
	err := app.DB.QueryRowContext(ctx, query, id).Scan(
		&l.ID, &l.Name, &l.Role, &l.Color, &l.CreatedAt,
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

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"lead": l}, nil)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// POST /leads
// CreateLead inserts a new lead and returns the generated row.
func (app *Application) CreateLead(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name  string `json:"name"`
		Role  string `json:"role"`
		Color string `json:"color"`
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
INSERT INTO leads (name, role, color)
VALUES ($1, $2, $3)
RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var l models.Lead
	l.Name = input.Name
	l.Role = input.Role
	l.Color = input.Color

	err = app.DB.QueryRowContext(ctx, query, input.Name, input.Role, input.Color).Scan(
		&l.ID, &l.CreatedAt,
	)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	extra := http.Header{"Location": []string{"/leads/" + l.ID}}
	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"lead": l}, extra)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// PUT /leads/{id}
// UpdateLead replaces all fields on an existing lead row.
func (app *Application) UpdateLead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		helpers.NotFound(w)
		return
	}

	var input struct {
		Name  string `json:"name"`
		Role  string `json:"role"`
		Color string `json:"color"`
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
UPDATE leads
SET name = $1, role = $2, color = $3
WHERE id = $4`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result, err := app.DB.ExecContext(ctx, query, input.Name, input.Role, input.Color, id)
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

	updated := models.Lead{ID: id, Name: input.Name, Role: input.Role, Color: input.Color}
	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"lead": updated}, nil)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// DELETE /leads/{id}
// DeleteLead removes a lead row.
// Returns 204 No Content on success.
func (app *Application) DeleteLead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		helpers.NotFound(w)
		return
	}

	query := `DELETE FROM leads WHERE id = $1`

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
