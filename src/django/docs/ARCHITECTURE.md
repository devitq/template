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

## Why this pattern?
- **Separation of Concerns**: Clear distinction between API logic and business logic.
- **Testability**: Business logic can be tested independently of the HTTP layer.
- **Reusability**: Services can be called from multiple interfaces (API, Celery tasks, management commands).
- **Maintainability**: Smaller, focused functions are easier to understand and maintain.
