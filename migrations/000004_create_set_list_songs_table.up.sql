CREATE TABLE set_list_songs (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    set_list_id UUID NOT NULL REFERENCES set_lists(id) ON DELETE CASCADE,
    song_id UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    section TEXT,
    note TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(set_list_id, song_id)
);