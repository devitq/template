from django.contrib.auth.models import AbstractUser
from django.utils.translation import gettext_lazy as _

from apps.core.models import BaseModel


class User(AbstractUser, BaseModel):
    class Meta:
        swappable = "AUTH_USER_MODEL"
        verbose_name = _("user")  # type: ignore[mutable-override]
        verbose_name_plural = _("users")  # type: ignore[mutable-override]
