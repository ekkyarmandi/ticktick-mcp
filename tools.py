import httpx
from mcp.server.fastmcp import FastMCP
from config import TICKTICK_API_BASE, HEADERS

mcp = FastMCP("ticktick")


@mcp.tool()
async def get_projects():
    """
    Returns a list of projects from TickTick.
    """
    url = f"{TICKTICK_API_BASE}/project"
    async with httpx.AsyncClient() as client:
        response = await client.get(url, headers=HEADERS)
        return response.json()


@mcp.tool()
async def project_details(project_id: str):
    """
    Returns details of a specific project in TickTick.
    """
    url = f"{TICKTICK_API_BASE}/project/{project_id}/data"
    async with httpx.AsyncClient() as client:
        response = await client.get(url, headers=HEADERS)
        return response.json()


@mcp.tool()
async def get_task_details(project_id: str, task_id: str):
    """
    Returns details of a specific task in TickTick.
    """
    url = f"{TICKTICK_API_BASE}/project/{project_id}/task/{task_id}"
    async with httpx.AsyncClient() as client:
        response = await client.get(url, headers=HEADERS)
        return response.json()


@mcp.tool()
async def create_project(project_name: str):
    """
    Creates a new project in TickTick.

    Body Args:
    name, name of the project
    """
    url = f"{TICKTICK_API_BASE}/project"
    payload = {"name": project_name}
    async with httpx.AsyncClient() as client:
        response = await client.post(url, headers=HEADERS, json=payload)
        return response.json()


@mcp.tool()
async def create_task(
    project_id: str,
    title: str,
    content: str | None = None,
    desc: str | None = None,
    isAllDay: bool | None = None,
    startDate: str | None = None,
    dueDate: str | None = None,
    timeZone: str | None = None,
    reminders: list | None = None,
    repeatFlag: str | None = None,
    priority: int | None = 0,
    sortOrder: int | None = None,
    items: list | None = None,
):
    """
    Creates a new task in TickTick.

    Args:
        project_id: ID of the project the task belongs to.
        title: Task title (required).
        content: Task content.
        desc: Description of checklist.
        isAllDay: All-day event flag.
        startDate: Start date/time (e.g., "2019-11-13T03:00:00+0000").
        dueDate: Due date/time (e.g., "2019-11-13T03:00:00+0000").
        timeZone: Time zone for the dates.
        reminders: List of reminders.
        repeatFlag: Recurrence rules.
        priority: Task priority (0=Normal, higher=higher priority).
        sortOrder: Sort order.
        items: List of subtasks (dicts with keys like 'title', 'startDate', etc.).
    """
    url = f"{TICKTICK_API_BASE}/task"
    payload = {
        "projectId": project_id,
        "title": title,
    }
    if content is not None:
        payload["content"] = content
    if desc is not None:
        payload["desc"] = desc
    if isAllDay is not None:
        payload["isAllDay"] = isAllDay
    if startDate is not None:
        payload["startDate"] = startDate
    if dueDate is not None:
        payload["dueDate"] = dueDate
    if timeZone is not None:
        payload["timeZone"] = timeZone
    if reminders is not None:
        payload["reminders"] = reminders
    if repeatFlag is not None:
        payload["repeatFlag"] = repeatFlag
    if priority is not None:
        payload["priority"] = priority
    if sortOrder is not None:
        payload["sortOrder"] = sortOrder
    if items is not None:
        payload["items"] = items

    if isAllDay is False and startDate is None:
        pass

    async with httpx.AsyncClient() as client:
        response = await client.post(url, headers=HEADERS, json=payload)
        return response.json()


@mcp.tool()
async def update_task(
    task_id: str,
    project_id: str,
    title: str | None = None,
    content: str | None = None,
    desc: str | None = None,
    isAllDay: bool | None = None,
    startDate: str | None = None,
    dueDate: str | None = None,
    timeZone: str | None = None,
    reminders: list | None = None,
    repeatFlag: str | None = None,
    priority: int | None = None,
    sortOrder: int | None = None,
    items: list | None = None,
):
    """
    Updates a task in TickTick.

    Args:
        task_id: Task identifier (required, for URL path).
        project_id: Project ID (required, in body).
        title: Task title.
        content: Task content.
        desc: Description of checklist.
        isAllDay: All-day event flag.
        startDate: Start date/time (e.g., "2019-11-13T03:00:00+0000").
        dueDate: Due date/time (e.g., "2019-11-13T03:00:00+0000").
        timeZone: Time zone for the dates.
        reminders: List of reminders.
        repeatFlag: Recurrence rules.
        priority: Task priority (0=Normal, higher=higher priority).
        sortOrder: Sort order.
        items: List of subtasks (dicts with keys like 'title', 'startDate', etc.).
    """
    url = f"{TICKTICK_API_BASE}/task/{task_id}"
    payload = {
        "id": task_id,
        "projectId": project_id,
    }
    if title is not None:
        payload["title"] = title
    if content is not None:
        payload["content"] = content
    if desc is not None:
        payload["desc"] = desc
    if isAllDay is not None:
        payload["isAllDay"] = isAllDay
    if startDate is not None:
        payload["startDate"] = startDate
    if dueDate is not None:
        payload["dueDate"] = dueDate
    if timeZone is not None:
        payload["timeZone"] = timeZone
    if reminders is not None:
        payload["reminders"] = reminders
    if repeatFlag is not None:
        payload["repeatFlag"] = repeatFlag
    if priority is not None:
        payload["priority"] = priority
    if sortOrder is not None:
        payload["sortOrder"] = sortOrder
    if items is not None:
        payload["items"] = items

    if isAllDay is False and startDate is None:
        pass

    async with httpx.AsyncClient() as client:
        response = await client.post(url, headers=HEADERS, json=payload)
        return response.json()


@mcp.tool()
async def complete_task(task_id: str):
    """
    Completes a task in TickTick.
    """
    url = f"{TICKTICK_API_BASE}/task/{task_id}/complete"
    async with httpx.AsyncClient() as client:
        response = await client.post(url, headers=HEADERS)
        return response.json()


@mcp.tool()
async def delete_task(task_id: str):
    """
    Deletes a task in TickTick.
    """
    url = f"{TICKTICK_API_BASE}/task/{task_id}"
    async with httpx.AsyncClient() as client:
        response = await client.delete(url, headers=HEADERS)
        return response.json()
