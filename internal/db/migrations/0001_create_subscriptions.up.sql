CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS subscriptions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  service_name TEXT NOT NULL,
  price INTEGER NOT NULL,
  user_id UUID NOT NULL,
  start_month INTEGER NOT NULL,
  start_year INTEGER NOT NULL,
  end_month INTEGER,
  end_year INTEGER,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
