package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/Jkicker-bz/systems_test_1/internal/handlers"
	_ "github.com/lib/pq"
)

// application holds the database connection pool.
//
// A *sql.DB is NOT a single connection — it is a managed pool that opens and
// closes connections automatically as demand fluctuates. You create it once at
// startup, store it here, and pass it into the handlers via the Application struct.
// This is the "dependency injection via struct" pattern used throughout this codebase.
type application struct {
	db *sql.DB
}

func main() {

	dsn := "postgres://worship_app_test:1234@localhost:5432/worship_app_test?sslmode=disable"

	db, err := openDB(dsn)
	if err != nil {
		log.Fatalf("cannot open database: %v", err)
	}
	defer db.Close()

	// Pass the DB pool into the handlers Application struct so all
	// handlers have access to it via their method receiver.
	app := &handlers.Application{DB: db}

	mux := http.NewServeMux()

	// ── Leads ────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /leads", app.ListLeads)
	mux.HandleFunc("GET /leads/{id}", app.GetLead)
	mux.HandleFunc("POST /leads", app.CreateLead)
	mux.HandleFunc("PUT /leads/{id}", app.UpdateLead)
	mux.HandleFunc("DELETE /leads/{id}", app.DeleteLead)

	// ── Songs ─────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /songs", app.ListSongs)
	mux.HandleFunc("GET /songs/{id}", app.GetSong)
	mux.HandleFunc("POST /songs", app.CreateSong)
	mux.HandleFunc("PUT /songs/{id}", app.UpdateSong)
	mux.HandleFunc("DELETE /songs/{id}", app.DeleteSong)

	// ── Set Lists ─────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /set_lists", app.ListSetLists)
	mux.HandleFunc("GET /set_lists/{id}", app.GetSetList)
	mux.HandleFunc("POST /set_lists", app.CreateSetList)
	mux.HandleFunc("PUT /set_lists/{id}", app.UpdateSetList)
	mux.HandleFunc("DELETE /set_lists/{id}", app.DeleteSetList)

	// ── Set List Songs ────────────────────────────────────────────────────────
	mux.HandleFunc("GET /set_lists/{id}/songs", app.ListSetListSongs)
	mux.HandleFunc("GET /set_lists/{id}/songs/{song_id}", app.GetSetListSong)
	mux.HandleFunc("POST /set_lists/{id}/songs", app.AddSongToSetList)
	mux.HandleFunc("DELETE /set_lists/{id}/songs/{song_id}", app.RemoveSongFromSetList)

	// ── Performances ──────────────────────────────────────────────────────────
	mux.HandleFunc("GET /performances", app.ListPerformances)
	mux.HandleFunc("GET /performances/{id}", app.GetPerformance)
	mux.HandleFunc("POST /performances", app.CreatePerformance)
	mux.HandleFunc("DELETE /performances/{id}", app.DeletePerformance)

	log.Println("Starting server on :4000")
	log.Println()
	log.Println("  GET    /leads")
	log.Println("  GET    /leads/{id}")
	log.Println("  POST   /leads")
	log.Println("  PUT    /leads/{id}")
	log.Println("  DELETE /leads/{id}")
	log.Println("  GET    /songs")
	log.Println("  GET    /songs/{id}")
	log.Println("  POST   /songs")
	log.Println("  PUT    /songs/{id}")
	log.Println("  DELETE /songs/{id}")
	log.Println("  GET    /set_lists")
	log.Println("  GET    /set_lists/{id}")
	log.Println("  POST   /set_lists")
	log.Println("  PUT    /set_lists/{id}")
	log.Println("  DELETE /set_lists/{id}")
	log.Println("  GET    /set_lists/{id}/songs")
	log.Println("  GET    /set_lists/{id}/songs/{song_id}")
	log.Println("  POST   /set_lists/{id}/songs")
	log.Println("  DELETE /set_lists/{id}/songs/{song_id}")
	log.Println("  GET    /performances")
	log.Println("  GET    /performances/{id}")
	log.Println("  POST   /performances")
	log.Println("  DELETE /performances/{id}")

	err = http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxIdleTime(15 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	log.Println("Database connection pool established")
	return db, nil
}
