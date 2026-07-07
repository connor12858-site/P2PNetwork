from fastapi import APIRouter, Query

from auth import verify
from models import BootstrapNode
from services import registry

router = APIRouter()


@router.post("/register")
def register(
    node: BootstrapNode,
    password: str = Query(...)
):

    verify(password)

    registry.register(node)

    return {"success": True}


@router.delete("/register/{peer_id}")
def unregister(
    peer_id: str,
    password: str = Query(...)
):

    verify(password)

    registry.unregister(peer_id)

    return {"success": True}