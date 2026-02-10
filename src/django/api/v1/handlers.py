import logging
from collections.abc import Callable
from http import HTTPStatus as status
from typing import Any

import django.core.exceptions
import django.http
import ninja.errors
from django.http import HttpRequest, HttpResponse
from django.utils.timezone import now
from django_guid import get_guid
from ninja import NinjaAPI

from api.v1.schemas import ApiError, FieldError, ValidationError
from config.errors import ConflictError, ForbiddenError

logger = logging.getLogger("django")


def create_error_response(
    request: HttpRequest,
    code: str,
    message: str,
    http_status: int,
    router: NinjaAPI,
    details: dict[str, Any] | None = None,
) -> HttpResponse:
    error_data = ApiError(
        code=code,
        message=message,
        traceId=str(get_guid()),
        timestamp=now(),
        path=request.path,
        details=details,
    )
    return router.create_response(request, error_data, status=http_status)


def handle_validation_error(
    request: HttpRequest,
    exc: ninja.errors.ValidationError,
    router: NinjaAPI,
) -> HttpResponse:
    field_errors: list[FieldError] = []
    for error in exc.errors:
        # Ninja/Pydantic v2 error format
        loc = error.get("loc", [])
        field = ".".join(map(str, loc)) if loc else "non_field_error"
        field_errors.append(
            FieldError(
                field=field,
                issue=error.get("msg", "Unknown error"),
                rejectedValue=error.get("input"),
            )
        )

    error_data = ValidationError(
        code="VALIDATION_FAILED",
        message="Validation failed",
        traceId=str(get_guid()),
        timestamp=now(),
        path=request.path,
        fieldErrors=field_errors,
    )
    return router.create_response(request, error_data, status=status.BAD_REQUEST)


def handle_django_validation_error(
    request: HttpRequest,
    exc: django.core.exceptions.ValidationError,
    router: NinjaAPI,
) -> HttpResponse:
    field_errors: list[FieldError] = []
    if hasattr(exc, "error_dict"):
        for field, errors in exc.error_dict.items():
            field_errors.extend(
                FieldError(
                    field=field,
                    issue=str(error.message),
                    rejectedValue=None,
                )
                for error in errors
            )
    else:
        field_errors.extend(
            FieldError(
                field="non_field_error",
                issue=str(error.message),
                rejectedValue=None,
            )
            for error in exc.error_list
        )

    error_data = ValidationError(
        code="VALIDATION_FAILED",
        message="Validation failed",
        traceId=str(get_guid()),
        timestamp=now(),
        path=request.path,
        fieldErrors=field_errors,
    )
    return router.create_response(request, error_data, status=status.BAD_REQUEST)


def handle_authentication_error(
    request: HttpRequest,
    exc: ninja.errors.AuthenticationError,
    router: NinjaAPI,
) -> HttpResponse:
    return create_error_response(
        request,
        code="UNAUTHENTICATED",
        message="Authentication required",
        http_status=status.UNAUTHORIZED,
        router=router,
    )


def handle_forbidden_error(
    request: HttpRequest,
    exc: ForbiddenError,
    router: NinjaAPI,
) -> HttpResponse:
    return create_error_response(
        request,
        code="FORBIDDEN",
        message=exc.message,
        http_status=status.FORBIDDEN,
        router=router,
    )


def handle_not_found_error(
    request: HttpRequest,
    exc: Exception,
    router: NinjaAPI,
) -> HttpResponse:
    return create_error_response(
        request,
        code="NOT_FOUND",
        message="Resource not found",
        http_status=status.NOT_FOUND,
        router=router,
    )


def handle_conflict_error(
    request: HttpRequest,
    exc: ConflictError,
    router: NinjaAPI,
) -> HttpResponse:
    # Reuse handle_django_validation_error but change status code
    response = handle_django_validation_error(request, exc.validation_error, router)
    response.status_code = status.CONFLICT
    return response


def handle_unknown_exception(
    request: HttpRequest,
    exc: Exception,
    router: NinjaAPI,
) -> HttpResponse:
    logger.error("Internal server error: %s", exc, exc_info=True)  # noqa: LOG014

    return create_error_response(
        request,
        code="INTERNAL_SERVER_ERROR",
        message="An unexpected error occurred",
        http_status=status.INTERNAL_SERVER_ERROR,
        router=router,
    )


exception_handlers: list[tuple[Any, Callable[..., Any]]] = [
    (ninja.errors.ValidationError, handle_validation_error),
    (django.core.exceptions.ValidationError, handle_django_validation_error),
    (ninja.errors.AuthenticationError, handle_authentication_error),
    (ForbiddenError, handle_forbidden_error),
    (django.http.Http404, handle_not_found_error),
    (ConflictError, handle_conflict_error),
    (Exception, handle_unknown_exception),
]
