from fastapi import APIRouter

from repositories.registry_repo import count

router = APIRouter()


@router.get("/health")
def health():

    return {
        "status": "ok",
        "registered_nodes": count()
    }