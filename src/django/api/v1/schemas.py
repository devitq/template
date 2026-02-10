from http import HTTPStatus as status

from ninja import Schema


class BadRequestError(Schema):
    detail: str = status.BAD_REQUEST.phrase


class UnauthorizedError(Schema):
    detail: str = status.UNAUTHORIZED.phrase


class ForbiddenError(Schema):
    detail: str = status.FORBIDDEN.phrase


class NotFoundError(Schema):
    detail: str = status.NOT_FOUND.phrase


class ConflictError(Schema):
    detail: str = status.CONFLICT.phrase
