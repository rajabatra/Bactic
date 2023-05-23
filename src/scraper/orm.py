from typing import List, Optional
from enum import Enum
from datetime import datetime
from sqlalchemy import String, ForeignKey, CheckConstraint
from sqlalchemy.orm import Mapped, DeclarativeBase, relationship, mapped_column

class Base(DeclarativeBase):
    pass

class Sex(Enum):
    MALE = 0
    FEMALE = 1

class EventType(Enum):
    _100m = 0
    _200m = 1
    _400m = 2
    _800m = 3
    _1500m = 4
    _5000m = 5
    _10000m = 6
    high_jump = 2
    

class EventStage(Enum):
    PRELIM = 0
    FINAL = 1

class Athlete(Base):
    __tablename__ = "athlete"

    id: Mapped[int] = mapped_column(primary_key = True)
    name: Mapped[str] = mapped_column(String(30))
    year: Mapped[int]
    school_id = mapped_column(ForeignKey("school.id"))
    sex: Mapped[Sex]

class School(Base):
    __tablename__ = "school"

    id: Mapped[int] = mapped_column(primary_key=True)
    division: Mapped[int]
    name: Mapped[str] = mapped_column(String(30))
    conference: Mapped[Optional[str]] = mapped_column(String(30))

class Meet(Base):
    __tablename__ = "meet"
    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(30))
    date: Mapped[datetime]

class Result(Base):
    __tablename__ = "result"
    id: Mapped[int] = mapped_column(primary_key=True)
    athlete_id: Mapped[int] = mapped_column(ForeignKey('athlete.id'))
    event_type: Mapped[EventType]
    time: Mapped[Optional[float]] = mapped_column(CheckConstraint("""
    ((event_type = '_100m' or
        event_type = '_200m' or
        event_type = '_400m' or
        event_type = '_800m' or
        event_type = '_1500m' or 
        event_type = '_5000m' or
        event_type = '_10000m')
        and time is not null)
        """))
    height: Mapped[Optional[float]] = mapped_column(CheckConstraint("""
    (event_type = 'vault' or
    event_type = 'jump' or
    event_type = 'triple_jump or
    )"""))

    def __init__(self, athlete_id: int, event_id: int, time: float=None, place: int=None, height: float=None, wind: float=None):
        self.athlete_id = athlete_id
        self.event_id = event_id
        self.time = time
        self.place = place
        self.height = height
        self.wind = wind
        

    def __rep__(self):
        return f'{self.id}, {self.time}, {self.height}'

class Event(Base):
    __tablename__ = "event"
    id: Mapped[int] = mapped_column(primary_key=True)
    type: Mapped[EventType]
    event_stage: Mapped[Optional[EventStage]]
    date: Mapped[datetime] # just make sure that this is correct
    
    def __init__(self, type: EventType, date: datetime, event_stage: EventStage=None):
        self.type = type
        self.date = date
        self.event_stage = event_stage
