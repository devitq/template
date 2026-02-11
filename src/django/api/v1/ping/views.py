import logging
from http import HTTPStatus as status

from django.http import HttpRequest
from ninja import Router

from api.v1 import schemas as global_schemas
from api.v1.ping import schemas

# Standard way to get a logger in Django views
logger = logging.getLogger("template")

router = Router(tags=["ping"])


@router.get(
    "",
    response={
        status.OK: schemas.PingOut,
        status.BAD_REQUEST: global_schemas.ValidationError,
        status.NOT_FOUND: global_schemas.ApiError,
    },
)
def ping(request: HttpRequest) -> tuple[status, schemas.PingOut]:
    # Example of using the logger
    logger.info("Ping endpoint called")
    return status.OK, schemas.PingOut()
