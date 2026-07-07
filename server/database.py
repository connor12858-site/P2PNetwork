import sqlite3

from config import CONFIG

DB = CONFIG["db"]


def get_connection():
    conn = sqlite3.connect(DB)
    conn.row_factory = sqlite3.Row
    return conn


def initialize():

    conn = get_connection()

    conn.execute("""
        CREATE TABLE IF NOT EXISTS bootstrap_nodes (

            peer_id TEXT PRIMARY KEY,

            address TEXT NOT NULL,
            port INTEGER NOT NULL,

            updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP

        )
    """)

    conn.commit()
    conn.close()