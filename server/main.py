from fastapi import FastAPI, HTTPException, Query
import yaml
import json

from Model import BootstrapNode

# Load config
def load_config():
    with open("./data/config.yaml", "r") as file:
        return yaml.safe_load(file)

config = load_config()


# Read nodes safely
def get_registered_nodes():
    try:
        with open(config["bs-path"], "r") as file:
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

    nodes = get_registered_nodes()

    # remove duplicates by peer_id
    nodes = [n for n in nodes if n.get("peer_id") != node.peer_id]
    nodes.append(node.model_dump())

    with open(config["bs-path"], "w") as file:
        for n in nodes:
            file.write(json.dumps(n) + "\n")

    return {"success": True}


# Unregister node
@app.delete("/register/{peer_id}")
def unregister(
    peer_id: str,
    password: str = Query(...)
):
    if password != auth_credentials["password"]:
        raise HTTPException(status_code=401, detail="Unauthorized")

    nodes = get_registered_nodes()

    updated = [n for n in nodes if n.get("peer_id") != peer_id]

    with open(config["bs-path"], "w") as file:
        for n in updated:
            file.write(json.dumps(n) + "\n")

    return {"success": True}


# Health
@app.get("/health")
def health():
    return {
        "status": "ok",
        "registered_nodes": len(get_registered_nodes())
    }