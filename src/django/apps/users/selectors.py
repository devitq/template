from django.db.models import QuerySet

from apps.users.models import User


def user_get_by_id(user_id: str) -> User | None:
    return User.objects.filter(id=user_id).first()


def user_list() -> QuerySet[User]:
    return User.objects.all()
