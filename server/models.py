from pydantic import BaseModel


class BootstrapNode(BaseModel):
    peer_id: str
    address: str
    port: int | None = None
    name: str