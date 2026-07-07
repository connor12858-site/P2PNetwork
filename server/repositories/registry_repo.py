from database import get_connection
import re


def _resolve_port(node):
    if node.port is not None:
        return node.port

    match = re.search(r"/tcp/(\d+)(?:/|$)", node.address)
    if match:
        return int(match.group(1))

    raise ValueError("bootstrap node port is required")


def all_nodes():

    conn = get_connection()

    rows = conn.execute(
        "SELECT peer_id,address,port FROM bootstrap_nodes"
    ).fetchall()

    conn.close()

    return [dict(r) for r in rows]


def register(node):

    conn = get_connection()

    port = _resolve_port(node)

    conn.execute("""
        INSERT INTO bootstrap_nodes(peer_id,address,port)
        VALUES(?,?,?)
        ON CONFLICT(peer_id)
        DO UPDATE SET
            address=excluded.address,
            port=excluded.port,
            updated=CURRENT_TIMESTAMP
    """, (
        node.peer_id,
        node.address,
        port
    ))

    conn.commit()
    conn.close()


def unregister(peer):

    conn = get_connection()

    conn.execute(
        "DELETE FROM bootstrap_nodes WHERE peer_id=?",
        (peer,)
    )

    conn.commit()
    conn.close()


def count():

    conn = get_connection()

    c = conn.execute(
        "SELECT COUNT(*) FROM bootstrap_nodes"
    ).fetchone()[0]

    conn.close()

    return c