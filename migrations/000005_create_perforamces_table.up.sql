CREATE TABLE performances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    song_id UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    set_list_id UUID REFERENCES set_lists(id) ON DELETE SET NULL,
    lead_id UUID REFERENCES leads(id) ON DELETE SET NULL,
    key_played TEXT,
    date_played DATE,
    event TEXT, 
    note TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);  