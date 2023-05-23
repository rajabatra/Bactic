DROP TYPE IF EXISTS sex;
CREATE TYPE sex as ENUM('male', 'female');

DROP TYPE IF EXISTS event_type;
CREATE TYPE event_type as ENUM('5000m');

CREATE TABLE IF NOT EXISTS athletes(
    id INT NOT NULL PRIMARY KEY,
    name VARCHAR NOT NULL,
    year INT NOT NULL,
    school_id INT NOT NULL,
    sex sex NOT NULL
);

CREATE TABLE IF NOT EXISTS schools(
    id INT NOT NULL PRIMARY KEY,
    division INT NOT NULL,
    name VARCHAR NOT NULL,
    conference VARCHAR 
);

CREATE TABLE IF NOT EXISTS events(
    id INT NOT NULL PRIMARY KEY,
    type event_type NOT NULL,
    name VARCHAR NOT NULL,
    date DATE NOT NULL,
    meet_id INT NOT NULL
);

CREATE TABLE IF NOT EXISTS meets(
    id INT NOT NULL PRIMARY KEY,
    name VARCHAR NOT NULL,
    date DATE NOT NULL
);

CREATE TABLE IF NOT EXISTS 5000m(
    id INT NOT NULL PRIMARY KEY,
    athlete_id INT NOT NULL,
    race_id INT NOT NULL,
    time INT NOT NULL,
    place INT NOT NULL
);