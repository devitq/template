from http import HTTPStatus as status

from django.http import HttpRequest, JsonResponse

from config.utils import build_error_payload


def create_error_response(
    request: HttpRequest,
    code: str,
    message: str,
    http_status: int,
) -> JsonResponse:
    payload = build_error_payload(request, code, message)
    # JsonResponse expects a dict, but build_error_payload returns datetime for timestamp
    # We should ensure it's serializable if not using a custom encoder,
    # but JsonResponse handles datetime by default if using Django's encoder.
    return JsonResponse(status=http_status, data=payload)


def handler400(
    request: HttpRequest,
    exception: Exception | None = None,
) -> JsonResponse:
    return create_error_response(
        request,
        code="BAD_REQUEST",
        message=status.BAD_REQUEST.phrase,
        http_status=status.BAD_REQUEST,
    )


def handler403(
    request: HttpRequest,
    exception: Exception | None = None,
) -> JsonResponse:
    return create_error_response(
        request,
        code="FORBIDDEN",
        message=status.FORBIDDEN.phrase,
        http_status=status.FORBIDDEN,
    )


def handler404(
    request: HttpRequest,
    exception: Exception | None = None,
) -> JsonResponse:
    return create_error_response(
        request,
        code="NOT_FOUND",
        message=status.NOT_FOUND.phrase,
        http_status=status.NOT_FOUND,
    )


def handler500(
    request: HttpRequest,
    exception: Exception | None = None,
) -> JsonResponse:
    return create_error_response(
        request,
        code="INTERNAL_SERVER_ERROR",
        message=status.INTERNAL_SERVER_ERROR.phrase,
        http_status=status.INTERNAL_SERVER_ERROR,
    )
