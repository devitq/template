from ninja import Schema


class PingOut(Schema):
    detail: str = "pong"
