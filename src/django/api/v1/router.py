from functools import partial
from typing import Any, override

import orjson
from django.http import HttpRequest
from ninja import NinjaAPI, Schema
from ninja.renderers import BaseRenderer

from api.v1 import handlers
from api.v1.ping.views import router as ping_router


class ORJSONRenderer(BaseRenderer):
    media_type: str | None = "application/json"

    @override
    def render(self, request: HttpRequest, data: Any, *, response_status: int) -> Any:
        return orjson.dumps(data, default=self.default)

    def default(self, obj: Any) -> Any:
        if isinstance(obj, Schema):
            return obj.model_dump(by_alias=True)
        raise TypeError


router = NinjaAPI(
    title="Template API",
    version="1",
    description="API docs for Template",
    openapi_url="/docs/openapi.json",
    renderer=ORJSONRenderer(),
)


router.add_router(
    "ping",
    ping_router,
)


for exception, handler in handlers.exception_handlers:
    router.add_exception_handler(exception, partial(handler, router=router))
