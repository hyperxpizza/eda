CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS messages (
    id uuid DEFAULT uuid_generate_v4(),
    content TEXT NOT NULL,
    PRIMARY KEY (id)
);