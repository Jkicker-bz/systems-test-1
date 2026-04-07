CREATE TABLE leads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    role TEXT,
    color TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);  