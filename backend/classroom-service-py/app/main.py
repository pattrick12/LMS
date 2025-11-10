from fastapi import FastAPI
from app.routers import classrooms, assignments

app = FastAPI(
    title="LMS Classroom Service",
    description="Manages classroom content, assignments, and submissions.",
)

@app.get("/health", tags=["Health"])
async def health_check():
    """Health check endpoint."""
    return {"status": "classroom service is up and running"}

# Include all the API routers
app.include_router(classrooms.router)
app.include_router(assignments.router)

