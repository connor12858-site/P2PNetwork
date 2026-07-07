import json
from pathlib import Path
from threading import Lock

import yaml
from fastapi import FastAPI, HTTPException, Query

from Model import BootstrapNode

BASE_DIR = Path(__file__).resolve().parent
CONFIG_PATH = BASE_DIR / "data" / "config.yaml"
REGISTRY_LOCK = Lock()


def load_config() -> dict:
    with CONFIG_PATH.open("r", encoding="utf-8") as file:
        config = yaml.safe_load(file)

    if not isinstance(config, dict):
        raise RuntimeError("Invalid server config format")

    return config


def resolve_registry_path() -> Path:
    registry_path = Path(str(config["bs-path"]))
    if registry_path.is_absolute():
        return registry_path
    return (BASE_DIR / registry_path).resolve()


config = load_config()
REGISTRY_PATH = resolve_registry_path()


def get_registered_nodes():
    if not REGISTRY_PATH.exists():
        return []

    try:
        with REGISTRY_PATH.open("r", encoding="utf-8") as file:
            nodes = []
            for line in file:
                line = line.strip()
                if not line:
                    continue
                nodes.append(json.loads(line))
            return nodes
    except FileNotFoundError:
        return []
    except json.JSONDecodeError:
        return []


def write_registered_nodes(nodes):
    REGISTRY_PATH.parent.mkdir(parents=True, exist_ok=True)

    temp_path = REGISTRY_PATH.with_suffix(REGISTRY_PATH.suffix + ".tmp")
    with temp_path.open("w", encoding="utf-8") as file:
        for node in nodes:
            file.write(json.dumps(node) + "\n")

    temp_path.replace(REGISTRY_PATH)


# endpoints list
endpoints = {
    "[GET] /": "Gets all registered nodes",
    "[POST] /register": "Registers a bootstrap node",
    "[DELETE] /register/{peer_id}": "Removes a bootstrap node",
    "[GET] /health": "Health check",
    "[GET] /help": "List endpoints"
}

auth_credentials = {
    "password": config["password"]
}

app = FastAPI()


# Get all nodes
@app.get("/")
def root():
    return {"nodes": get_registered_nodes()}


# Help
@app.get("/help")
def help():
    return endpoints


# Register node
@app.post("/register")
def register(
    node: BootstrapNode,
    password: str = Query(...)
):
    if password != auth_credentials["password"]:
        raise HTTPException(status_code=401, detail="Unauthorized")

    with REGISTRY_LOCK:
        nodes = get_registered_nodes()

        nodes = [n for n in nodes if n.get("peer_id") != node.peer_id]
        nodes.append(node.model_dump())

        write_registered_nodes(nodes)

    return {"success": True}


# Unregister node
@app.delete("/register/{peer_id}")
def unregister(
    peer_id: str,
    password: str = Query(...)
):
    if password != auth_credentials["password"]:
        raise HTTPException(status_code=401, detail="Unauthorized")

    with REGISTRY_LOCK:
        nodes = get_registered_nodes()

        updated = [n for n in nodes if n.get("peer_id") != peer_id]

        write_registered_nodes(updated)

    return {"success": True}


# Health
@app.get("/health")
def health():
    return {
        "status": "ok",
        "registered_nodes": len(get_registered_nodes())
    }