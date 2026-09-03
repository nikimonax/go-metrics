CREATE TABLE metrics (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR NOT NULL,
    type VARCHAR NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION,

    CONSTRAINT metrics_chk_type CHECK (type IN ('gauge', 'counter')),
    CONSTRAINT metrics_chk_delta CHECK (delta IS NULL OR delta >= 0),
    CONSTRAINT metrics_chk_value CHECK (value IS NULL or value >= 0),
    CONSTRAINT metrics_chk_type_and_value CHECK (
        (type = 'counter' AND delta IS NOT NULL AND value IS NULL) OR
        (type = 'gauge' AND value IS NOT NULL AND delta IS NULL)
    ),
    CONSTRAINT metrics_unique_name_and_type UNIQUE (name, type)
);
