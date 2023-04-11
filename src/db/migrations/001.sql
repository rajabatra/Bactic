CREATE TABLE times(
    INT NOT NULL PRIMARY KEY id,
    INT NOT NULL athlete_id,
    INT NOT NULL race_id,
    INT NOT NULL time,
    INT NOT NULL place
);

CREATE TABLE athletes(
    INT NOT NULL PRIMARY KEY id,
    VARCHAR NOT NULL name,
    INT NOT NULL year,
    INT NOT NULL school_id
);

CREATE TABLE schools(
    INT NOT NULL PRIMARY KEY id,
    INT NOT NULL division,
    VARCHAR NOT NULL name,
    VARCHAR NOT NULL conference 
);

CREATE TABLE races(
    INT NOT NULL PRIMARY KEY id,
    VARCHAR NOT NULL name,
    DATE NOT NULL date
);