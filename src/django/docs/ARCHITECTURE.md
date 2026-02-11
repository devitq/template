# Architecture

This project follows a layered architecture pattern, specifically a "Service Layer" approach.

## Layers

### 1. Interfaces (API)
Located in `api/`. This layer is responsible for handling HTTP requests, parsing input, and returning responses. It uses Django Ninja.
It should NOT contain business logic. It calls **Services** or **Selectors**.

### 2. Services
Located in `apps/<app_name>/services.py`. Services contain business logic and side effects (e.g., creating objects, sending emails).
They are responsible for maintaining data integrity and performing actions.

### 3. Selectors
Located in `apps/<app_name>/selectors.py`. Selectors are responsible for fetching data from the database.
They return QuerySets or model instances.

### 4. Domain (Models)
Located in `apps/<app_name>/models.py`. Standard Django models.

## Logging

Logging is pre-configured and integrated with standard Django logging and OpenTelemetry.

### Usage in Views/Services

To use logging, obtain a logger instance for the project:

```python
import logging

# Use the pre-defined project logger name
logger = logging.getLogger("template")

def my_view(request):
    logger.info("Informational message")
    try:
        # ...
    except Exception as e:
        logger.error("Something went wrong: %s", e, exc_info=True)
```

### Configuration
- In **development** (`DEBUG=True`), logs are output in a human-readable, colored format.
- In **production** (`DEBUG=False`), logs are output in JSON format, suitable for log aggregators, and include a `correlation_id` (Trace ID) for request tracing.

## Why this pattern?
- **Separation of Concerns**: Clear distinction between API logic and business logic.
- **Testability**: Business logic can be tested independently of the HTTP layer.
- **Reusability**: Services can be called from multiple interfaces (API, Celery tasks, management commands).
- **Maintainability**: Smaller, focused functions are easier to understand and maintain.
