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

// GET /performances
// ListPerformances fetches every performance ordered by date descending.
// COALESCE converts any NULL column into an empty string so the
// plain string fields in models.Performance are always safe to scan into.
func (app *Application) ListPerformances(w http.ResponseWriter, r *http.Request) {
	query := `
SELECT id, song_id,
COALESCE(set_list_id::text, ''),
COALESCE(lead_id::text, ''),
COALESCE(key_played, ''),
COALESCE(date_played::text, ''),
COALESCE(event, ''),
COALESCE(note, ''),
created_at
FROM performances
ORDER BY date_played DESC`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := app.DB.QueryContext(ctx, query)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	defer rows.Close()

	var performances []models.Performance

	for rows.Next() {
		var p models.Performance
		err := rows.Scan(
			&p.ID, &p.SongID, &p.SetListID, &p.LeadID,
			&p.KeyPlayed, &p.DatePlayed, &p.Event, &p.Note, &p.CreatedAt,
		)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		performances = append(performances, p)
	}

	if err = rows.Err(); err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"performances": performances}, nil)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// GET /performances/{id}
// GetPerformance fetches a single performance by primary key.
func (app *Application) GetPerformance(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		helpers.NotFound(w)
		return
	}

	query := `
SELECT id, song_id,
COALESCE(set_list_id::text, ''),
COALESCE(lead_id::text, ''),
COALESCE(key_played, ''),
COALESCE(date_played::text, ''),
COALESCE(event, ''),
COALESCE(note, ''),
created_at
FROM performances
WHERE id = $1`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var p models.Performance
	err := app.DB.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.SongID, &p.SetListID, &p.LeadID,
		&p.KeyPlayed, &p.DatePlayed, &p.Event, &p.Note, &p.CreatedAt,
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

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"performance": p}, nil)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// POST /performances
// CreatePerformance logs a new performance.
func (app *Application) CreatePerformance(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SongID     string `json:"song_id"`
		SetListID  string `json:"set_list_id"`
		LeadID     string `json:"lead_id"`
		KeyPlayed  string `json:"key_played"`
		DatePlayed string `json:"date_played"`
		Event      string `json:"event"`
		Note       string `json:"note"`
	}

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		helpers.BadRequest(w, err.Error())
		return
	}

	v := helpers.NewValidator()
	v.Check(input.SongID != "", "song_id", "must be provided")
	v.Check(input.DatePlayed != "", "date_played", "must be provided")

	if !v.Valid() {
		helpers.FailedValidation(w, v.Errors)
		return
	}

	query := `
INSERT INTO performances (song_id, set_list_id, lead_id, key_played, date_played, event, note)
VALUES ($1, NULLIF($2, '')::uuid, NULLIF($3, '')::uuid, $4, $5::date, $6, $7)
RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var p models.Performance
	p.SongID = input.SongID
	p.SetListID = input.SetListID
	p.LeadID = input.LeadID
	p.KeyPlayed = input.KeyPlayed
	p.DatePlayed = input.DatePlayed
	p.Event = input.Event
	p.Note = input.Note

	err = app.DB.QueryRowContext(ctx, query,
		input.SongID, input.SetListID, input.LeadID, input.KeyPlayed,
		input.DatePlayed, input.Event, input.Note,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	extra := http.Header{"Location": []string{"/performances/" + p.ID}}
	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"performance": p}, extra)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// DELETE /performances/{id}
// DeletePerformance removes a performance log entry.
// Returns 204 No Content on success.
func (app *Application) DeletePerformance(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		helpers.NotFound(w)
		return
	}

	query := `DELETE FROM performances WHERE id = $1`

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
