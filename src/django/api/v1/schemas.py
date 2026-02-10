from datetime import datetime
from typing import Any

from ninja import Schema
from pydantic import ConfigDict, Field


class FieldError(Schema):
    model_config = ConfigDict(populate_by_name=True)

    field: str = Field(..., description="Field name with error (can be nested, e.g. 'location.latitude')")
    issue: str = Field(..., description="Problem description")
    rejected_value: Any = Field(None, alias="rejectedValue", description="Value that failed validation")


class ApiError(Schema):
    model_config = ConfigDict(populate_by_name=True)

    code: str
    message: str
    trace_id: str = Field(..., alias="traceId")
    timestamp: datetime
    path: str
    details: dict[str, Any] | None = None


class ValidationError(ApiError):
    field_errors: list[FieldError] = Field(..., alias="fieldErrors")
