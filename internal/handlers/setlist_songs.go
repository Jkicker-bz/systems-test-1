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

// GET /set_lists/{id}/songs
// ListSetListSongs fetches all songs in a set list ordered by order_index.
// COALESCE converts any NULL column into an empty string/zero so the
// plain fields in models.SetListSong are always safe to scan into.
func (app *Application) ListSetListSongs(w http.ResponseWriter, r *http.Request) {
	setListID := r.PathValue("id")
	if setListID == "" {
		helpers.NotFound(w)
		return
	}

	query := `
SELECT id, set_list_id, song_id,
COALESCE(order_index, 0),
COALESCE(key_override, ''),
COALESCE(section, ''),
COALESCE(note, ''),
created_at
FROM set_list_songs
WHERE set_list_id = $1
ORDER BY order_index`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := app.DB.QueryContext(ctx, query, setListID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	defer rows.Close()

	var setListSongs []models.SetListSong

	for rows.Next() {
		var sls models.SetListSong
		err := rows.Scan(
			&sls.ID, &sls.SetListID, &sls.SongID, &sls.OrderIndex,
			&sls.KeyOverride, &sls.Section, &sls.Note, &sls.CreatedAt,
		)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		setListSongs = append(setListSongs, sls)
	}

	if err = rows.Err(); err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"set_list_songs": setListSongs}, nil)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// GET /set_lists/{id}/songs/{song_id}
// GetSetListSong fetches a single song entry within a set list.
func (app *Application) GetSetListSong(w http.ResponseWriter, r *http.Request) {
	setListID := r.PathValue("id")
	songID := r.PathValue("song_id")
	if setListID == "" || songID == "" {
		helpers.NotFound(w)
		return
	}

	query := `
SELECT id, set_list_id, song_id,
COALESCE(order_index, 0),
COALESCE(key_override, ''),
COALESCE(section, ''),
COALESCE(note, ''),
created_at
FROM set_list_songs
WHERE set_list_id = $1 AND song_id = $2`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var sls models.SetListSong
	err := app.DB.QueryRowContext(ctx, query, setListID, songID).Scan(
		&sls.ID, &sls.SetListID, &sls.SongID, &sls.OrderIndex,
		&sls.KeyOverride, &sls.Section, &sls.Note, &sls.CreatedAt,
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

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"set_list_song": sls}, nil)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// POST /set_lists/{id}/songs
// AddSongToSetList adds a song to a set list.
func (app *Application) AddSongToSetList(w http.ResponseWriter, r *http.Request) {
	setListID := r.PathValue("id")
	if setListID == "" {
		helpers.NotFound(w)
		return
	}

	var input struct {
		SongID      string `json:"song_id"`
		OrderIndex  int    `json:"order_index"`
		KeyOverride string `json:"key_override"`
		Section     string `json:"section"`
		Note        string `json:"note"`
	}

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		helpers.BadRequest(w, err.Error())
		return
	}

	v := helpers.NewValidator()
	v.Check(input.SongID != "", "song_id", "must be provided")

	if !v.Valid() {
		helpers.FailedValidation(w, v.Errors)
		return
	}

	query := `
INSERT INTO set_list_songs (set_list_id, song_id, order_index, key_override, section, note)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var sls models.SetListSong
	sls.SetListID = setListID
	sls.SongID = input.SongID
	sls.OrderIndex = input.OrderIndex
	sls.KeyOverride = input.KeyOverride
	sls.Section = input.Section
	sls.Note = input.Note

	err = app.DB.QueryRowContext(ctx, query,
		setListID, input.SongID, input.OrderIndex, input.KeyOverride, input.Section, input.Note,
	).Scan(&sls.ID, &sls.CreatedAt)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"set_list_song": sls}, nil)
	if err != nil {
		helpers.ServerError(w, err)
	}
}

// DELETE /set_lists/{id}/songs/{song_id}
// RemoveSongFromSetList removes a song from a set list.
// Returns 204 No Content on success.
func (app *Application) RemoveSongFromSetList(w http.ResponseWriter, r *http.Request) {
	setListID := r.PathValue("id")
	songID := r.PathValue("song_id")
	if setListID == "" || songID == "" {
		helpers.NotFound(w)
		return
	}

	query := `DELETE FROM set_list_songs WHERE set_list_id = $1 AND song_id = $2`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result, err := app.DB.ExecContext(ctx, query, setListID, songID)
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
