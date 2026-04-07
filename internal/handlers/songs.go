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

// GET /songs
// ListSongs fetches every song row ordered by title.
// COALESCE converts any NULL column into an empty string/zero so the
// plain string fields in models.Song are always safe to scan into.
func (app *Application) ListSongs(w http.ResponseWriter, r *http.Request) {
	query := `
SELECT id, title,
COALESCE(artist, ''),
COALESCE(style, ''),
COALESCE(key, ''),
COALESCE(bpm, 0),
COALESCE(chord_sheet, ''),
COALESCE(yt_link, ''),
COALESCE(color, ''),
COALESCE(note, ''),
created_at
FROM songs
ORDER BY title`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := app.DB.QueryContext(ctx, query)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	defer rows.Close()

	var songs []models.Song

	for rows.Next() {
		var s models.Song
		err := rows.Scan(
			&s.ID, &s.Title, &s.Artist, &s.Style, &s.Key,
			&s.BPM, &s.ChordSheet, &s.YTLink, &s.Color, &s.Note, &s.CreatedAt,
		)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		songs = append(songs, s)
	}

	if err = rows.Err(); err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"songs": songs}, nil)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// GET /songs/{id}
// GetSong fetches a single song by primary key.
func (app *Application) GetSong(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		helpers.NotFound(w)
		return
	}

	query := `
SELECT id, title,
COALESCE(artist, ''),
COALESCE(style, ''),
COALESCE(key, ''),
COALESCE(bpm, 0),
COALESCE(chord_sheet, ''),
COALESCE(yt_link, ''),
COALESCE(color, ''),
COALESCE(note, ''),
created_at
FROM songs
WHERE id = $1`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var s models.Song
	err := app.DB.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.Title, &s.Artist, &s.Style, &s.Key,
		&s.BPM, &s.ChordSheet, &s.YTLink, &s.Color, &s.Note, &s.CreatedAt,
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

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"song": s}, nil)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// POST /songs
// CreateSong inserts a new song and returns the generated row.
func (app *Application) CreateSong(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title      string `json:"title"`
		Artist     string `json:"artist"`
		Style      string `json:"style"`
		Key        string `json:"key"`
		BPM        int    `json:"bpm"`
		ChordSheet string `json:"chord_sheet"`
		YTLink     string `json:"yt_link"`
		Color      string `json:"color"`
		Note       string `json:"note"`
	}

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		helpers.BadRequest(w, err.Error())
		return
	}

	v := helpers.NewValidator()
	v.Check(input.Title != "", "title", "must be provided")

	if !v.Valid() {
		helpers.FailedValidation(w, v.Errors)
		return
	}

	query := `
INSERT INTO songs (title, artist, style, key, bpm, chord_sheet, yt_link, color, note)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var s models.Song
	s.Title = input.Title
	s.Artist = input.Artist
	s.Style = input.Style
	s.Key = input.Key
	s.BPM = input.BPM
	s.ChordSheet = input.ChordSheet
	s.YTLink = input.YTLink
	s.Color = input.Color
	s.Note = input.Note

	err = app.DB.QueryRowContext(ctx, query,
		input.Title, input.Artist, input.Style, input.Key, input.BPM,
		input.ChordSheet, input.YTLink, input.Color, input.Note,
	).Scan(&s.ID, &s.CreatedAt)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	extra := http.Header{"Location": []string{"/songs/" + s.ID}}
	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"song": s}, extra)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// PUT /songs/{id}
// UpdateSong replaces all fields on an existing song row.
func (app *Application) UpdateSong(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		helpers.NotFound(w)
		return
	}

	var input struct {
		Title      string `json:"title"`
		Artist     string `json:"artist"`
		Style      string `json:"style"`
		Key        string `json:"key"`
		BPM        int    `json:"bpm"`
		ChordSheet string `json:"chord_sheet"`
		YTLink     string `json:"yt_link"`
		Color      string `json:"color"`
		Note       string `json:"note"`
	}

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		helpers.BadRequest(w, err.Error())
		return
	}

	v := helpers.NewValidator()
	v.Check(input.Title != "", "title", "must be provided")

	if !v.Valid() {
		helpers.FailedValidation(w, v.Errors)
		return
	}

	query := `
UPDATE songs
SET title = $1, artist = $2, style = $3, key = $4, bpm = $5,
   chord_sheet = $6, yt_link = $7, color = $8, note = $9
WHERE id = $10`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result, err := app.DB.ExecContext(ctx, query,
		input.Title, input.Artist, input.Style, input.Key, input.BPM,
		input.ChordSheet, input.YTLink, input.Color, input.Note, id,
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

	updated := models.Song{
		ID: id, Title: input.Title, Artist: input.Artist, Style: input.Style,
		Key: input.Key, BPM: input.BPM, ChordSheet: input.ChordSheet,
		YTLink: input.YTLink, Color: input.Color, Note: input.Note,
	}
	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"song": updated}, nil)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// DELETE /songs/{id}
// DeleteSong removes a song row.
// Returns 204 No Content on success.
func (app *Application) DeleteSong(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		helpers.NotFound(w)
		return
	}

	query := `DELETE FROM songs WHERE id = $1`

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
