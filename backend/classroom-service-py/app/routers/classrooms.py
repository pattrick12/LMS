from fastapi import APIRouter, Depends, HTTPException, status, Body
from typing import List
from app.models import Classroom, ClassroomIn, Module, Announcement
from app.security import get_current_user, get_instructor_or_ta, User
from app.db import classroom_collection, PyObjectId
from bson import ObjectId

router = APIRouter(
    prefix="/classrooms",
    tags=["Classrooms"]
)


@router.get("/me", response_model=List[Classroom])
async def get_my_classrooms(user: User = Depends(get_current_user)):
    """Fetches all classrooms a user is enrolled in (as student, TA, or instructor)."""
    query = {
        "$or": [
            {"student_ids": user.id},
            {"ta_ids": user.id},
            {"instructor_id": user.id}
        ]
    }
    classrooms = await classroom_collection.find(query).to_list(100)
    return classrooms


@router.post("/sync", status_code=status.HTTP_200_OK)
async def sync_classroom_from_erp(
        course_id: str = Body(...),
        instructor_id: str = Body(...),
        semester: str = Body(...),
        name: str = Body(...),
        student_ids: List[str] = Body(...)
):
    """
    (Gateway-Internal) Idempotently creates or updates a classroom.
    This is the core of "automatic enrollment". It should ONLY be called by the gateway.
    """
    query = {"course_id": course_id, "semester": semester}

    # We update the student list, name, and instructor every time.
    # We only set the other fields if the document is being created (upsert=True)
    update = {
        "$set": {
            "instructor_id": instructor_id,
            "name": name,
            "student_ids": student_ids
        },
        "$setOnInsert": {
            "ta_ids": [],
            "announcements": [],
            "modules": []
        }
    }

    await classroom_collection.update_one(query, update, upsert=True)
    return {"message": "Classroom synced successfully"}


@router.post("/{classroom_id}/modules", response_model=Module)
async def add_module(
        classroom_id: str,
        module: Module,
        user: User = Depends(get_instructor_or_ta)  # RBAC check
):
    """(Instructor/TA) Post a new lecture or module."""
    if not ObjectId.is_valid(classroom_id):
        raise HTTPException(status.HTTP_400_BAD_REQUEST, "Invalid Classroom ID")

    # Check that the user is actually the instructor or TA for *this* class
    update_result = await classroom_collection.update_one(
        {"_id": ObjectId(classroom_id), "$or": [{"instructor_id": user.id}, {"ta_ids": user.id}]},
        {"$push": {"modules": module.model_dump()}}
    )
    if update_result.matched_count == 0:
        raise HTTPException(status.HTTP_403_FORBIDDEN, "Not authorized or classroom not found")

    return module


@router.post("/{classroom_id}/announcements", response_model=Announcement)
async def add_announcement(
        classroom_id: str,
        announcement: Announcement,
        user: User = Depends(get_instructor_or_ta)  # RBAC check
):
    """(Instructor/TA) Post a new announcement."""
    if not ObjectId.is_valid(classroom_id):
        raise HTTPException(status.HTTP_400_BAD_REQUEST, "Invalid Classroom ID")

    update_result = await classroom_collection.update_one(
        {"_id": ObjectId(classroom_id), "$or": [{"instructor_id": user.id}, {"ta_ids": user.id}]},
        {"$push": {"announcements": announcement.model_dump()}}
    )
    if update_result.matched_count == 0:
        raise HTTPException(status.HTTP_403_FORBIDDEN, "Not authorized or classroom not found")

    return announcement