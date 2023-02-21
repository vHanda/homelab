create database homelab;
-- \c homelab;

CREATE TABLE IF NOT EXISTS living_room (
   time        TIMESTAMPTZ       NOT NULL,
   temperature DOUBLE PRECISION  NOT NULL,
   humidity    DOUBLE PRECISION  NOT NULL
);

SELECT create_hypertable('living_room', 'time', if_not_exists => TRUE);

ALTER TABLE living_room SET (timescaledb.compress);
SELECT add_compression_policy('living_room', INTERVAL '7 days', if_not_exists => TRUE);