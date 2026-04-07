CREATE TABLE songs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    artist TEXT,
    style TEXT,
    key TEXT,
    bpm INTEGER,
    chord_sheet TEXT,
    yt_link TEXT,
    color TEXT,
    note TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);