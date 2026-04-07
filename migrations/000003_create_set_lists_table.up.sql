CREATE TABLE set_lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    event_type TEXT,
    date DATE,
    notes TEXT,
    created_by UUID REFERENCES leads(id) ON DELETE SET NULL, 
    created_at TIMESTAMP NOT NULL DEFAULT NOW ()
);