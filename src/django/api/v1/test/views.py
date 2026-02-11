from http import HTTPStatus as status

from django.http import HttpRequest
from ninja import Router

from api.v1 import schemas as global_schemas
from api.v1.test import schemas

router = Router(tags=["test"])


@router.post(
    "",
    response={
        status.OK: schemas.TestSchemaOut,
        status.BAD_REQUEST: global_schemas.ValidationError,
        status.NOT_FOUND: global_schemas.ApiError,
    },
)
def test(request: HttpRequest, test: schemas.TestSchemaIn) -> tuple[status, schemas.TestSchemaOut]:
    return status.OK, schemas.TestSchemaOut(id=1)
