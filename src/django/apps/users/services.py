from typing import Any

from apps.users.models import User


def user_create(*, username: str, email: str, password: str | None = None, **extra_fields: Any) -> User:
    user = User(username=username, email=email, **extra_fields)
    if password:
        user.set_password(password)
    else:
        user.set_unusable_password()

    user.save()
    return user
