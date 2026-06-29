"""
Configuration for STAROps SDK
STAROps SDK 配置
"""

import os
from dataclasses import dataclass
from typing import Optional

from dotenv import load_dotenv

from .errors import SDKException


@dataclass
class Config:
    """应用配置 / Application configuration"""
    workspace: str
    endpoint: str
    access_key_id: str
    access_key_secret: str
    region: str = "cn-hangzhou"
    employee_name: str = "default"

    @classmethod
    def load_from_env(cls) -> "Config":
        """从环境变量加载配置 / Load configuration from environment variables"""
        load_dotenv()

        workspace = os.getenv("VIBEOPS_WORKSPACE", "")
        endpoint = os.getenv("VIBEOPS_ENDPOINT", "")
        region = os.getenv("VIBEOPS_REGION", "cn-hangzhou")
        access_key_id = os.getenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "")
        access_key_secret = os.getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "")
        employee_name = os.getenv("VIBEOPS_EMPLOYEE_NAME", "default")

        # Validate required fields
        missing_vars = []
        if not endpoint:
            missing_vars.append("VIBEOPS_ENDPOINT")
        if not access_key_id:
            missing_vars.append("ALIBABA_CLOUD_ACCESS_KEY_ID")
        if not access_key_secret:
            missing_vars.append("ALIBABA_CLOUD_ACCESS_KEY_SECRET")

        if missing_vars:
            raise SDKException.config_missing(missing_vars)

        return cls(
            workspace=workspace,
            endpoint=endpoint,
            access_key_id=access_key_id,
            access_key_secret=access_key_secret,
            region=region or "cn-hangzhou",
            employee_name=employee_name or "default",
        )
