CREATE TABLE IF NOT EXISTS athlete(
    id BIGINT PRIMARY KEY,
    name VARCHAR,
    year INT
);

CREATE TABLE IF NOT EXISTS result(
    id BIGINT PRIMARY KEY,
    heat_id BIGINT,
    ath_id BIGINT,
    pl SMALLINT,
    quant FLOAT,
    wind_ms FLOAT,
    stage TINYINT,
    FOREIGN KEY(heat_id) REFERENCES heat(id),
    FOREIGN KEY(ath_id) REFERENCES athlete(id)
);

CREATE TABLE IF NOT EXISTS heat(
    id BIGINT PRIMARY KEY,
    meet_id BIGINT,
    event_type TINYINT,
    FOREIGN KEY(meet_id) REFERENCES meet(id)
);

CREATE TABLE IF NOT EXISTS school(
    id BIGINT PRIMARY KEY,
    name VARCHAR,
    division TINYINT,
    url VARCHAR NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS meet(
    id BIGINT PRIMARY KEY,
    name VARCHAR,
    date DATE
);

CREATE TABLE IF NOT EXISTS league(
    school_id BIGINT NOT NULL,
    league_name VARCHAR NOT NULL,
    FOREIGN KEY(school_id) REFERENCES school(id)
);

CREATE TABLE IF NOT EXISTS athlete_in_school(
    athlete_id BIGINT NOT NULL,
    school_id BIGINT NOT NULL,
    FOREIGN KEY(athlete_id) REFERENCES athlete(id),
    FOREIGN KEY(school_id) REFERENCES school(id),
    PRIMARY KEY(athlete_id, school_id)
);
