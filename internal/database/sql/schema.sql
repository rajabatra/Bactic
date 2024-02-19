CREATE TABLE IF NOT EXISTS athlete(
    id BIGINT PRIMARY KEY,
    name VARCHAR NOT NULL,
    year INT,
    zscore FLOAT
);

INSERT INTO athlete(id, name) VALUES(0, 'NULL') ON CONFLICT DO NOTHING; -- Null for results without an athlete profile

CREATE TABLE IF NOT EXISTS meet(
    id BIGINT PRIMARY KEY,
    name VARCHAR NOT NULL,
    season SMALLINT NOT NULL,
    date DATE NOT NULL
);

CREATE TABLE IF NOT EXISTS heat(
    id BIGINT PRIMARY KEY,
    meet_id BIGINT,
    event_type SMALLINT NOT NULL,
    FOREIGN KEY(meet_id) REFERENCES meet(id)
);

CREATE TABLE IF NOT EXISTS result(
    id BIGINT PRIMARY KEY,
    heat_id BIGINT,
    ath_id BIGINT,
    pl SMALLINT,
    quant FLOAT,
    wind_ms FLOAT,
    stage SMALLINT,
    FOREIGN KEY(heat_id) REFERENCES heat(id),
    FOREIGN KEY(ath_id) REFERENCES athlete(id)
);

CREATE TABLE IF NOT EXISTS school(
    id BIGINT PRIMARY KEY,
    name VARCHAR NOT NULL,
    division SMALLINT NOT NULL,
    url VARCHAR NOT NULL UNIQUE
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

CREATE TABLE IF NOT EXISTS athlete_map(
    x BIGINT PRIMARY KEY,
    y BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS distns(
    event SMALLINT PRIMARY KEY,
    mean FLOAT NOT NULL,
    var FLOAT NOT NULL
);
