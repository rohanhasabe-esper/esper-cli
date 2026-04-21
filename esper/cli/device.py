"""
Device commands — replaces esper/controllers/device/device.py.
`espercli device list/show/set-active/unset-active`
"""
import uuid
from typing import Optional

import typer
from esperclient.rest import ApiException

from esper.cli.completions import device_state_complete, group_name_complete, gms_complete
from esper.cli.output import render
from esper.cli.state import state, validate_creds, parse_error_message
from esper.controllers.enums import DeviceState, OutputFormat
from esper.ext.api_client import APIClient
from esper.ext.db_wrapper import DBWrapper

app = typer.Typer(help="Device commands")


# ---------------------------------------------------------------------------
# Shared helpers
# ---------------------------------------------------------------------------

def _get_name_and_tags(device):
    name = device.alias_name if device.alias_name else device.device_name
    tags = ", ".join(device.tags) if device.tags else ""
    return name, tags


def _device_basic_response(device, fmt=OutputFormat.TABULATED):
    valid_keys = ["id", "device_name", "alias_name", "suid", "api_level", "template_name", "is_gms"]
    current_state = DeviceState(device.status).name

    if fmt == OutputFormat.JSON:
        renderable = {k: v for k, v in device.to_dict().items() if k in valid_keys}
        renderable["state"] = current_state
        renderable["tags"] = device.tags
    else:
        title, details = "TITLE", "DETAILS"
        renderable = [{title: k, details: v} for k, v in device.to_dict().items() if k in valid_keys]
        renderable.append({title: "state", details: current_state})
        _, tags = _get_name_and_tags(device)
        renderable.append({title: "tags", details: tags})
    return renderable


# ---------------------------------------------------------------------------
# Commands
# ---------------------------------------------------------------------------

@app.command("list")
def device_list(
    state_filter: Optional[str] = typer.Option(
        None, "--state", "-s",
        help="Device state: active, inactive, disabled",
        shell_complete=device_state_complete,
    ),
    name: Optional[str] = typer.Option(None, "--name", "-n", help="Device name"),
    group: Optional[str] = typer.Option(
        None, "--group", "-g", help="Group name",
        shell_complete=group_name_complete,
    ),
    imei: Optional[str] = typer.Option(None, "--imei", "-im", help="IMEI number"),
    serial: Optional[str] = typer.Option(None, "--serial", "-se", help="Serial number"),
    tags: Optional[str] = typer.Option(None, "--tags", "-t", help="Tags"),
    search: Optional[str] = typer.Option(None, "--search", help="Search device name, alias or ID"),
    brand: Optional[str] = typer.Option(None, "--brand", "-b", help="Brand name"),
    gms: Optional[str] = typer.Option(
        None, "--gms", "-gm", help="GMS or not: true, false",
        shell_complete=gms_complete,
    ),
    limit: int = typer.Option(20, "--limit", "-l", help="Results per page"),
    offset: int = typer.Option(0, "--offset", "-i", help="Initial index"),
    all_results: bool = typer.Option(
        False, "--all", "-A",
        help="Fetch ALL devices, auto-paginating (ignores --limit / --offset).",
    ),
    json_output: bool = typer.Option(False, "--json", "-j", help="Render result in JSON format"),
):
    """List devices in the enterprise."""
    validate_creds()
    db = DBWrapper(state.creds)
    device_client = APIClient(db.get_configure()).get_device_api_client()
    enterprise_id = db.get_enterprise_id()

    kwargs = {}
    if state_filter:
        if state_filter.upper() not in [s.name for s in DeviceState]:
            # match 'active'/'inactive'/'disabled' shorthand
            mapping = {"active": "ACTIVE", "inactive": "INACTIVE", "disabled": "DISABLED"}
            key = mapping.get(state_filter.lower())
            if key:
                kwargs["state"] = DeviceState[key].value
        else:
            kwargs["state"] = DeviceState[state_filter.upper()].value

    if name:
        kwargs["name"] = name
    if group:
        kw = {"name": group}
        group_id = None
        try:
            group_client = APIClient(db.get_configure()).get_group_api_client()
            search_response = group_client.get_all_groups(enterprise_id, limit=1, offset=0, **kw)
            for g in search_response.results:
                if g.name == group:
                    group_id = g.id
                    break
        except ApiException as e:
            state.log.error(f"[device-list] Failed to list groups: {e}")
            render(f"ERROR: {parse_error_message(e)}")
            raise typer.Exit(1)
        kwargs["group"] = group_id or str(uuid.uuid4())
    if imei:
        kwargs["imei"] = imei
    if serial:
        kwargs["serial"] = serial
    if search:
        kwargs["search"] = search
    if tags:
        kwargs["tags"] = tags
    if brand:
        kwargs["brand"] = brand
    if gms:
        kwargs["is_gms"] = gms

    try:
        if all_results:
            from esper.cli.output import console as _console
            results = []
            page_size, cur_offset = 100, 0
            with _console.status("[dim]Fetching all devices…[/dim]", spinner="dots"):
                while True:
                    page = device_client.get_all_devices(
                        enterprise_id, limit=page_size, offset=cur_offset, **kwargs
                    )
                    results.extend(page.results)
                    if len(results) >= page.count or not page.results:
                        break
                    cur_offset += page_size
            total = len(results)

            class _FakeResponse:
                count = total
                def __init__(self, r):
                    self.results = r
            response = _FakeResponse(results)
        else:
            response = device_client.get_all_devices(enterprise_id, limit=limit, offset=offset, **kwargs)
    except ApiException as e:
        state.log.error(f"[device-list] Failed to list devices: {e}")
        render(f"ERROR: {parse_error_message(e)}")
        raise typer.Exit(1)

    render(f"Number of Devices: {response.count}")
    if not json_output:
        devices = []
        label = {"id": "ID", "name": "NAME", "model": "MODEL", "state": "CURRENT STATE", "tags": "TAGS"}
        for device in response.results:
            current_state = DeviceState(device.status).name
            dev_name, dev_tags = _get_name_and_tags(device)
            devices.append({
                label["id"]: device.id,
                label["name"]: dev_name,
                label["model"]: device.hardware_info.get("manufacturer"),
                label["state"]: current_state,
                label["tags"]: dev_tags,
            })
        render(devices, format=OutputFormat.TABULATED.value, headers="keys", tablefmt="plain")
    else:
        devices = []
        for device in response.results:
            current_state = DeviceState(device.status).name
            dev_name, _ = _get_name_and_tags(device)
            devices.append({
                "id": device.id,
                "device": dev_name,
                "model": device.hardware_info.get("manufacturer"),
                "state": current_state,
                "tags": device.tags,
            })
        render(devices, format=OutputFormat.JSON.value)


@app.command("show")
def device_show(
    device_name: str = typer.Argument(..., help="Device name to show details for"),
    active: bool = typer.Option(
        False, "--active", "-a", help="Set this device as the active device"
    ),
    json_output: bool = typer.Option(False, "--json", "-j", help="Render result in JSON format"),
):
    """Show device details (and optionally set as active)."""
    validate_creds()
    db = DBWrapper(state.creds)
    device_client = APIClient(db.get_configure()).get_device_api_client()
    enterprise_id = db.get_enterprise_id()

    kwargs = {"name": device_name}
    try:
        search_response = device_client.get_all_devices(enterprise_id, limit=1, offset=0, **kwargs)
        if not search_response.results:
            state.log.debug(f"[device-show] Device does not exist with name {device_name}")
            render(f"Device does not exist with name {device_name}")
            raise typer.Exit(1)
        response = search_response.results[0]
    except ApiException as e:
        state.log.error(f"[device-show] Failed to list devices: {e}")
        render(f"ERROR: {parse_error_message(e)}")
        raise typer.Exit(1)

    if active:
        dev_name, _ = _get_name_and_tags(response)
        db.set_device({"id": response.id, "name": dev_name})

    if not json_output:
        renderable = _device_basic_response(response)
        render(renderable, format=OutputFormat.TABULATED.value, headers="keys", tablefmt="plain")
    else:
        renderable = _device_basic_response(response, OutputFormat.JSON)
        render(renderable, format=OutputFormat.JSON.value)


@app.command("set-active")
def set_active(
    name: Optional[str] = typer.Option(None, "--name", "-n", help="Device name"),
    json_output: bool = typer.Option(False, "--json", "-j", help="Render result in JSON format"),
):
    """Set or show the active device."""
    validate_creds()
    db = DBWrapper(state.creds)
    device_client = APIClient(db.get_configure()).get_device_api_client()
    enterprise_id = db.get_enterprise_id()

    if name:
        kwargs = {"name": name}
        try:
            search_response = device_client.get_all_devices(enterprise_id, limit=1, offset=0, **kwargs)
            if not search_response.results:
                state.log.debug(f"[device-active] Device does not exist with name {name}")
                render(f"Device does not exist with name {name}")
                raise typer.Exit(1)
            response = search_response.results[0]
            dev_name, _ = _get_name_and_tags(response)
            db.set_device({"id": response.id, "name": dev_name})
        except ApiException as e:
            state.log.error(f"[device-active] Failed to list devices: {e}")
            render(f"ERROR: {parse_error_message(e)}")
            raise typer.Exit(1)
    else:
        device = db.get_device()
        if device is None or device.get("name") is None:
            state.log.debug("[device-active] There is no active device.")
            render("There is no active device.")
            raise typer.Exit(1)

        device_id = device.get("id")
        try:
            response = device_client.get_device_by_id(enterprise_id, device_id)
        except ApiException as e:
            state.log.error(f"[device-active] Failed to show active device: {e}")
            render(f"ERROR: {parse_error_message(e)}")
            raise typer.Exit(1)

    if not json_output:
        renderable = _device_basic_response(response)
        render(renderable, format=OutputFormat.TABULATED.value, headers="keys", tablefmt="plain")
    else:
        renderable = _device_basic_response(response, OutputFormat.JSON)
        render(renderable, format=OutputFormat.JSON.value)


@app.command("unset-active")
def unset_active():
    """Unset the active device."""
    validate_creds()
    db = DBWrapper(state.creds)

    device = db.get_device()
    if device is None or device.get("name") is None:
        state.log.debug("[device-active] There is no active device.")
        render("There is no active device.")
        raise typer.Exit(1)

    db.unset_device()
    state.log.debug(f"[device-active] Unset the active device {device.get('name')}")
    render(f"Unset the active device {device.get('name')}")
