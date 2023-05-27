from typing import Any, List, Optional
from enum import Enum
from datetime import datetime, timedelta
from sqlalchemy import String, ForeignKey, CheckConstraint
from sqlalchemy.orm import Mapped, DeclarativeBase, relationship, mapped_column


class Base(DeclarativeBase):
    pass

class Sex(Enum):
    MALE = 0
    FEMALE = 1

class EventStage(Enum):
    PRELIM = 0
    FINAL = 1

class Division(Enum):
    NCAADI = 0
    NCAADII = 1
    NCAADIII = 2
    NAIA = 3

class EventType(Enum):
    _5000m = 0
    _100m = 1
    _200m = 2
    _400m = 3
    #TODO: add rest of events

class Athlete(Base):
    __tablename__ = "athlete"

    id: Mapped[int] = mapped_column(primary_key = True)
    name: Mapped[str] = mapped_column(String(30))
    year: Mapped[int]
    school_id = mapped_column(ForeignKey("school.id"))
    sex: Mapped[Sex]

    def __init__(self, name: str, year: int, school_id: int, sex: Sex):
        self.name = name
        self.year = year
        self.school_id = school_id
        self.sex = sex
        
class School(Base):
    __tablename__ = "school"

    id: Mapped[int] = mapped_column(primary_key=True)
    division: Mapped[Division]
    name: Mapped[str] = mapped_column(String(30))
    conference: Mapped[Optional[str]] = mapped_column(String(30))

    def __init__(self, division: Division, name: str, conference: str = None):
        self.division = division
        self.name = name
        self.conference = conference

class Meet(Base):
    __tablename__ = "meet"
    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(30))
    date: Mapped[datetime]

class Result(Base):
    __tablename__ = "result"
    id: Mapped[int] = mapped_column(primary_key=True)
    athlete_id: Mapped[int] = mapped_column(ForeignKey('athlete.id'))
    type: Mapped[EventType]
    place: Mapped[int]
    date: Mapped[datetime]
    time: Mapped[Optional[timedelta]]
    wind: Mapped[Optional[float]]
    stage: Mapped[Optional[EventStage]]

    def __init__(self, athlete_id:int, type: EventType, place: int, date: datetime, time: timedelta = None, wind: float = None, stage: EventStage = None):
        self.athlete_id = athlete_id
        self.type = type
        self.place = place
        self.date = date

        # TODO: add more complex validation logic for individual events
        # if type in {}

class _100m(Base):
    __tablename__ = "100m"
    id: Mapped[int] = mapped_column(ForeignKey('result.id'), primary_key=True)
    athlete_id: Mapped[int] = mapped_column(ForeignKey('athlete.id'))
    time: Mapped[timedelta]
    wind: Mapped[float]
    date: Mapped[datetime]
    place: Mapped[int]
    stage: Mapped[Optional[EventStage]]


class _200m(Base):
    __tablename__ = "200m"
    id: Mapped[int] = mapped_column(ForeignKey('result.id'), primary_key=True)
    athlete_id: Mapped[int] = mapped_column(ForeignKey('athlete.id'))
    time: Mapped[timedelta]
    wind: Mapped[float]
    date: Mapped[datetime]
    place: Mapped[int]
    stage: Mapped[Optional[EventStage]]



class _400m(Base):
    __tablename__ = "400m"
    id: Mapped[int] = mapped_column(ForeignKey('result.id'), primary_key=True)
    athlete_id: Mapped[int] = mapped_column(ForeignKey('athlete.id'))
    time: Mapped[timedelta]
    date: Mapped[datetime]
    place: Mapped[int]
    stage: Mapped[Optional[EventStage]]



class _800m(Base):
    __tablename__ = "800m"
    id: Mapped[int] = mapped_column(ForeignKey('result.id'), primary_key=True)
    athlete_id: Mapped[int] = mapped_column(ForeignKey('athlete.id'))
    time: Mapped[timedelta]
    date: Mapped[datetime]
    place: Mapped[int]
    stage: Mapped[Optional[EventStage]]

    def __repr__(self) -> str:
        return f'800m(time:{self.time}, date:{self.date}, place:{self.place})'

class _5000m(Base):
    __tablename__ = "5000m"
    id: Mapped[int] = mapped_column(ForeignKey('result.id'), primary_key=True)
    athlete_id: Mapped[int] = mapped_column(ForeignKey('athlete.id'))
    time: Mapped[timedelta]
    date: Mapped[datetime]
    place: Mapped[int]

    def __repr__(self) -> str:
        return f'5000m(time:{self.time}, date:{self.date}, place:{self.place})'
    
    def __init__(self, athlete_id: int, time: timedelta, place: int, date: datetime):
        self.athlete_id = athlete_id
        self.time = time
        self.place = place
        self.date = date