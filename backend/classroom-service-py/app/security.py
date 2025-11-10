import os
from fastapi import Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer
from jose import jwt, JWTError
from pydantic import BaseModel
from typing import Optional

# This MUST match the secret in your Go services (.env file)
JWT_SECRET = os.getenv("JWT_SECRET")
if not JWT_SECRET:
    raise RuntimeError("JWT_SECRET environment variable is not set")

ALGORITHM = "HS256"
# This just tells FastAPI to look for an "Authorization: Bearer <token>" header
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="token")


class User(BaseModel):
    id: str  # This is the 'sub' claim
    role: str


async def get_current_user(token: str = Depends(oauth2_scheme)) -> User:
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Invalid or expired token",
        headers={"WWW-Authenticate": "Bearer"},
    )
    try:
        # Decode the JWT from the auth-service
        payload = jwt.decode(token, JWT_SECRET, algorithms=[ALGORITHM])

        # The Go service sets 'sub' (Subject) as the UserID and 'role' as the Role
        user_id = payload.get("sub")
        role = payload.get("role")

        if user_id is None or role is None:
            raise credentials_exception

        return User(id=user_id, role=role)

    except JWTError:
        raise credentials_exception


# --- RBAC Dependencies ---

def get_instructor_or_ta(user: User = Depends(get_current_user)):
    """Dependency for actions only instructors or TAs can perform."""
    if user.role not in ["instructor", "ta"]:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Instructor or TA access required"
        )
    return user


def get_student(user: User = Depends(get_current_user)):
    """Dependency for actions students can perform."""
    # TAs are also students, so they can submit work
    if user.role not in ["student", "ta"]:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Student or TA access required"
        )
    return user