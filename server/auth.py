from fastapi import HTTPException

from config import CONFIG


def verify(password: str):

    if password != CONFIG["password"]:
        raise HTTPException(
            status_code=401,
            detail="Unauthorized"
        )