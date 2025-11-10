import motor.motor_asyncio
from pydantic_core import CoreSchema, core_schema
from pydantic import GetCoreSchemaHandler
from bson import ObjectId
from typing import Any

# --- Pydantic + BSON ObjectId Helper ---
# This class allows Pydantic to validate MongoDB's ObjectIds
class PyObjectId(ObjectId):
    @classmethod
    def __get_pydantic_core_schema__(
        cls, source_type: Any, handler: GetCoreSchemaHandler
    ) -> CoreSchema:
        def validate(v):
            if not ObjectId.is_valid(v):
                raise ValueError("Invalid ObjectId")
            return ObjectId(v)

        return core_schema.json_or_python_schema(
            json_schema=core_schema.string_schema(),
            python_schema=core_schema.plain_validator_function(validate),
        )

# --- Database Connection ---
# Connects to the container from your xxx.sh script
MONGO_URL = "mongodb://lms-mongo:27017"
client = motor.motor_asyncio.AsyncIOMotorClient(MONGO_URL)
db = client.lms_classroom

# --- Collections ---
classroom_collection = db.classrooms
assignment_collection = db.assignments
submission_collection = db.submissions