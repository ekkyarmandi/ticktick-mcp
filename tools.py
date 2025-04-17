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
    task_name: str,
    content: str,
    desc: str,
    isAllDay: bool,
    startDate: str,
    dueDate: str,
):
    """
    Creates a new task in TickTick.

    Body Args:
    projectId, id of the project
    name, name of the task
    title, title of the task
    content, content of the task
    desc, description of the task
    isAllDay, whether the task is all day
    startDate, start date of the task if not all day
    dueDate, due date of the task if not all day
    """
    url = f"{TICKTICK_API_BASE}/task"
    payload = {
        "projectId": project_id,
        "name": task_name,
        "title": task_name,
        "content": content,
        "desc": desc,
        "isAllDay": isAllDay,
    }
    if not isAllDay:
        payload["startDate"] = startDate
        payload["dueDate"] = dueDate

    async with httpx.AsyncClient() as client:
        response = await client.post(url, headers=HEADERS, json=payload)
        return response.json()


@mcp.tool()
async def update_task(
    task_id: str,
    task_name: str,
    content: str,
    desc: str,
    isAllDay: bool,
    startDate: str,
    dueDate: str,
):
    """
    Updates a task in TickTick.

    Body Args:
    name, name of the task
    title, title of the task
    content, content of the task
    desc, description of the task
    isAllDay, whether the task is all day
    """
    url = f"{TICKTICK_API_BASE}/task/{task_id} "
    payload = {
        "name": task_name,
        "title": task_name,
        "content": content,
        "desc": desc,
        "isAllDay": isAllDay,
    }
    if not isAllDay:
        payload["startDate"] = startDate
        payload["dueDate"] = dueDate

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
