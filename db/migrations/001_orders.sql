CREATE TABLE IF NOT EXISTS orders (
  id uuid PRIMARY KEY,
  idempotency_key text UNIQUE NOT NULL,
  customer_id text NOT NULL,
  amount_cents integer NOT NULL CHECK (amount_cents BETWEEN 1 AND 99),
  status text NOT NULL CHECK (status IN ('accepted', 'processing', 'completed', 'failed')),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS queue_intents (
  order_id uuid PRIMARY KEY REFERENCES orders(id),
  enqueued_at timestamptz
);

CREATE TABLE IF NOT EXISTS schema_migrations (
  version text PRIMARY KEY,
  applied_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO schema_migrations(version) VALUES ('001') ON CONFLICT DO NOTHING;
