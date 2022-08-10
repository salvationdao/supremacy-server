ALTER TABLE
    features
ADD
    COLUMN IF NOT EXISTS globally_enabled bool NOT NULL DEFAULT false;