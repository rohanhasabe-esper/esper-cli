# Thin helpers for direct REST calls not covered by the esperclient SDK.
import requests


def api_get(environment: str, api_key: str, path: str, **params) -> dict:
    url = f"https://{environment}-api.esper.cloud/api{path}"
    resp = requests.get(url, headers={"Authorization": f"Bearer {api_key}"}, params=params, timeout=10)
    resp.raise_for_status()
    return resp.json()


def api_post(environment: str, api_key: str, path: str, body: dict) -> dict:
    url = f"https://{environment}-api.esper.cloud/api{path}"
    resp = requests.post(
        url,
        headers={"Authorization": f"Bearer {api_key}", "Content-Type": "application/json"},
        json=body,
        timeout=10,
    )
    resp.raise_for_status()
    return resp.json()


def api_delete(environment: str, api_key: str, path: str) -> int:
    url = f"https://{environment}-api.esper.cloud/api{path}"
    resp = requests.delete(url, headers={"Authorization": f"Bearer {api_key}"}, timeout=10)
    return resp.status_code
