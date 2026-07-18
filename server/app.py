from fastapi import FastAPI
import uvicorn

from database import initialize
from services.registry import get_nodes

from api.register import router as register_router
from api.health import router as health_router
from api.help import router as help_router

from config import CONFIG

initialize()

app = FastAPI()


@app.get("/")
def root():
    return {"nodes": get_nodes()}


app.include_router(register_router)
app.include_router(health_router)
app.include_router(help_router)

if __name__ == "__main__":
    uvicorn.run(app, host=CONFIG["host"], port=CONFIG["port"])