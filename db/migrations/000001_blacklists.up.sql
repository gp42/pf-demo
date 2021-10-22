CREATE TABLE IF NOT EXISTS blacklists (
  id SERIAL PRIMARY KEY,
  block_ip inet UNIQUE,
  block_timestamp timestamp
);
