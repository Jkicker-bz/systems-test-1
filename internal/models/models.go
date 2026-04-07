package models

import "time"

// Lead represents a worship team member.
type Lead struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// Song represents a song in the master catalog.
type Song struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Artist     string    `json:"artist"`
	Style      string    `json:"style"`
	Key        string    `json:"key"`
	BPM        int       `json:"bpm"`
	ChordSheet string    `json:"chord_sheet"`
	YTLink     string    `json:"yt_link"`
	Color      string    `json:"color"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

// SetList represents a named collection of songs for a specific event.
type SetList struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	EventType string    `json:"event_type"`
	Date      string    `json:"date"`
	Notes     string    `json:"notes"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// SetListSong represents a song entry within a set list.
type SetListSong struct {
	ID          string    `json:"id"`
	SetListID   string    `json:"set_list_id"`
	SongID      string    `json:"song_id"`
	OrderIndex  int       `json:"order_index"`
	KeyOverride string    `json:"key_override"`
	Section     string    `json:"section"`
	Note        string    `json:"note"`
	CreatedAt   time.Time `json:"created_at"`
}

// Performance represents a log entry of a song being played at an event.
type Performance struct {
	ID         string    `json:"id"`
	SongID     string    `json:"song_id"`
	SetListID  string    `json:"set_list_id"`
	LeadID     string    `json:"lead_id"`
	KeyPlayed  string    `json:"key_played"`
	DatePlayed string    `json:"date_played"`
	Event      string    `json:"event"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}
