from fastapi import APIRouter, Depends, HTTPException, status
from typing import List
from app.models import Assignment, AssignmentIn, Submission, SubmissionIn, GradeIn
from app.security import get_instructor_or_ta, get_student, User
from app.db import assignment_collection, submission_collection, classroom_collection
from bson import ObjectId

router = APIRouter(
    prefix="/assignments",
    tags=["Assignments & Grading"]
)


@router.post("/", response_model=Assignment)
async def create_assignment(
        assignment_in: AssignmentIn,
        user: User = Depends(get_instructor_or_ta)  # RBAC check
):
    """(Instructor/TA) Create a new assignment for a classroom."""

    if not ObjectId.is_valid(assignment_in.classroom_id):
        raise HTTPException(status.HTTP_400_BAD_REQUEST, "Invalid Classroom ID")
    classroom_oid = ObjectId(assignment_in.classroom_id)

    # Verify user is an instructor/TA for this specific classroom
    classroom = await classroom_collection.find_one(
        {"_id": classroom_oid, "$or": [{"instructor_id": user.id}, {"ta_ids": user.id}]}
    )
    if not classroom:
        raise HTTPException(status.HTTP_403_FORBIDDEN, "Not authorized or classroom not found")

    # Create assignment
    assignment = Assignment(
        **assignment_in.model_dump(exclude={"classroom_id"}),
        classroom_id=classroom_oid
    )
    new_assignment = await assignment_collection.insert_one(assignment.model_dump(by_alias=True))

    created_assignment = await assignment_collection.find_one({"_id": new_assignment.inserted_id})
    return created_assignment


@router.post("/submissions", response_model=Submission)
async def submit_assignment(
        submission_in: SubmissionIn,
        user: User = Depends(get_student)  # RBAC check
):
    """(Student/TA) Submit work for an assignment."""
    if not ObjectId.is_valid(submission_in.assignment_id):
        raise HTTPException(status.HTTP_400_BAD_REQUEST, "Invalid Assignment ID")

    assignment_oid = ObjectId(submission_in.assignment_id)

    # Check if student is actually in the classroom for this assignment
    assignment = await assignment_collection.find_one({"_id": assignment_oid})
    if not assignment:
        raise HTTPException(status.HTTP_404_NOT_FOUND, "Assignment not found")

    classroom = await classroom_collection.find_one(
        {"_id": assignment["classroom_id"], "$or": [{"student_ids": user.id}, {"ta_ids": user.id}]}
    )
    if not classroom:
        raise HTTPException(status.HTTP_403_FORBIDDEN, "Not enrolled in this classroom")

    # Create submission (or update if resubmitting)
    submission = Submission(
        assignment_id=assignment_oid,
        student_id=user.id,
        content=submission_in.content
    )

    # Use update_one with upsert=True to allow resubmissions
    await submission_collection.update_one(
        {"assignment_id": assignment_oid, "student_id": user.id},
        {"$set": submission.model_dump(by_alias=True, exclude={"id"})},
        upsert=True
    )

    # Find and return the created/updated document
    updated_doc = await submission_collection.find_one(
        {"assignment_id": assignment_oid, "student_id": user.id}
    )
    return updated_doc


@router.post("/submissions/{submission_id}/grade", response_model=Submission)
async def grade_submission(
        submission_id: str,
        grade: GradeIn,
        user: User = Depends(get_instructor_or_ta)  # RBAC check
):
    """(Instructor/TA) Grade a student's submission."""
    if not ObjectId.is_valid(submission_id):
        raise HTTPException(status.HTTP_400_BAD_REQUEST, "Invalid Submission ID")

    sub_oid = ObjectId(submission_id)

    # Get submission and its assignment
    submission = await submission_collection.find_one({"_id": sub_oid})
    if not submission:
        raise HTTPException(status.HTTP_404_NOT_FOUND, "Submission not found")

    assignment = await assignment_collection.find_one({"_id": submission["assignment_id"]})
    if not assignment:
        raise HTTPException(status.HTTP_440_NOT_FOUND, "Assignment not found")

    # Check if grader is authorized for the classroom
    classroom = await classroom_collection.find_one(
        {"_id": assignment["classroom_id"], "$or": [{"instructor_id": user.id}, {"ta_ids": user.id}]}
    )
    if not classroom:
        raise HTTPException(status.HTTP_403_FORBIDDEN, "Not authorized to grade this submission")

    # Update the grade
    update_result = await submission_collection.find_one_and_update(
        {"_id": sub_oid},
        {"$set": {"grade": grade.grade, "graded_by": user.id}},
        return_document=True
    )

    

    return update_result