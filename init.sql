CREATE TABLE IF NOT EXISTS battles (
    id SERIAL PRIMARY KEY,
    boss_id VARCHAR(50),
    result VARCHAR(20),
    duration INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rankings (
    id SERIAL PRIMARY KEY,
    battle_id INTEGER REFERENCES battles(id),
    nickname VARCHAR(50),
    class VARCHAR(20),
    total_damage INTEGER,
    incidents_solved INTEGER
);