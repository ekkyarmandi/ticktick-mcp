from decouple import config

TICKTICK_API_BASE = "https://api.ticktick.com/open/v1"
TICKTICK_API_KEY = config("TICKTICK_API_KEY")
HEADERS = {
    "Authorization": f"Bearer {TICKTICK_API_KEY}",
    "Content-Type": "application/json",
}
