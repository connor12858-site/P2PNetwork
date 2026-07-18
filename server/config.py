from pathlib import Path
import yaml
import dotenv
import os

BASE_DIR = Path(__file__).resolve().parent.parent / "data"
CONFIG_PATH = BASE_DIR / "config.yaml"


def load_config():

    with open(CONFIG_PATH, "r", encoding="utf8") as f:
        cfg = yaml.safe_load(f)

    if cfg.get(".env"):
        dotenv.load_dotenv()

        cfg["password"] = os.getenv("PASSWORD", cfg["password"])
        cfg["db"] = BASE_DIR / os.getenv("DB_PATH", cfg["db"])

    return cfg


CONFIG = load_config()