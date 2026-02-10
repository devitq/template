from django.urls import path
from health_check.views import HealthCheckView

from api.v1.router import router as api_v1_router

urlpatterns = [
    path("", api_v1_router.urls),
    # Health endpoint
    path("health", HealthCheckView.as_view(), name="health_check_home"),
]
