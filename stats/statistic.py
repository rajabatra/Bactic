import abc
from datetime import timedelta
import psycopg

class StatABC(abc.ABC):
    def __init__(self, db: psycopg.Cursor, interval: timedelta=None):
        self._interval = interval if interval != None else timedelta(0, 0, 0, 0, 0, 0, 0)

    @abc.abstractclassmethod
    def get(self, query: str=None):
        pass

    @abc.abstractclassmethod
    def compute(self, query: str=None):
        pass

    @abc.abstractclassmethod
    def describe(self) -> str:
        pass

    def interval(self) -> str:
        return self.interval
    
class Histogram(StatABC):
    def __init__(self, interval: timedelta=None):
        super().__init__(self, interval)

    def __get__(self, query: str=None):
        pass