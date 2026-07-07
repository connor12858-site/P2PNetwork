from fastapi import FastAPI

from database import initialize
from services.registry import get_nodes

from api.register import router as register_router
from api.health import router as health_router
from api.help import router as help_router

initialize()

app = FastAPI()


@app.get("/")
def root():
    return {"nodes": get_nodes()}


app.include_router(register_router)
app.include_router(health_router)
app.include_router(help_router)