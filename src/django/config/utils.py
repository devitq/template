import uuid
from typing import Any

from django.http import HttpRequest
from django.utils.timezone import now
from django_guid import get_guid


def build_error_payload(
    request: HttpRequest,
    code: str,
    message: str,
    details: dict[str, Any] | None = None,
    field_errors: list[dict[str, Any]] | None = None,
) -> dict[str, Any]:
    trace_id = get_guid() or str(uuid.uuid4())

    payload = {
        "code": code,
        "message": message,
        "traceId": str(trace_id),
        "timestamp": now(),
        "path": request.path,
        "details": details,
    }
    if field_errors is not None:
        payload["fieldErrors"] = field_errors
    return payload
