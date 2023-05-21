CREATE TABLE IF NOT EXISTS times(
    id INT NOT NULL PRIMARY KEY,
    athlete_id INT NOT NULL,
    race_id INT NOT NULL,
    time INT NOT NULL,
    place INT NOT NULL
);

DROP TYPE IF EXISTS sex;
CREATE TYPE sex as ENUM('male', 'female');

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

CREATE TABLE IF NOT EXISTS races(
    id INT NOT NULL PRIMARY KEY,
    name VARCHAR NOT NULL,
    date DATE NOT NULL,
    meet_id INT NOT NULL
);

CREATE TABLE IF NOT EXISTS meets(
    id INT NOT NULL PRIMARY KEY,
    name VARCHAR NOT NULL,
    date DATE NOT NULL
);