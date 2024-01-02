from api.main import FastAPI, HTTPException, Request
from databases import Database
from pydantic import BaseModel
import uvicorn
from typing import List

# Define the database URL
#DATABASE_URL = "postgres://postgres:pass@bactic_db:5432/bactic?sslmode=disable"
DATABASE_URL = 'postgresql://postgres:pass@localhost:5432/bactic'
# Initialize FastAPI app
app = FastAPI()

# Initialize the database
database = Database(DATABASE_URL)

# Pydantic models to represent the data
class Athlete(BaseModel):
    id: int
    name: str
    school_id: int

# Middleware for database connection management
@app.middleware("http")
async def db_session_middleware(request: Request, call_next):
    response = None
    try:
        await database.connect()
        response = await call_next(request)
    finally:
        await database.disconnect()
    return response

@app.get("/athlete-info/{athlete_name}", response_model=Athlete)
async def get_athlete_info(athlete_name: str):
    # Join the athlete and athlete_in_school tables to fetch the required information
    query = """
    SELECT athlete.id, athlete.name, athlete_in_school.school_id 
    FROM athlete 
    JOIN athlete_in_school ON athlete.id = athlete_in_school.athlete_id 
    WHERE athlete.name ILIKE :athlete_name
    """
    result = await database.fetch_one(query, values={"athlete_name": f"%{athlete_name}%"})
    if result is None:
        raise HTTPException(status_code=404, detail="Athlete not found")
    return result



# Run the application
if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)