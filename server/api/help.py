from fastapi import APIRouter

router = APIRouter()


@router.get("/help")
def help():

    return {
        "[GET] /": "Gets all nodes",
        "[POST] /register": "Registers node",
        "[DELETE] /register/{peer_id}": "Removes node",
        "[GET] /health": "Health",
    }