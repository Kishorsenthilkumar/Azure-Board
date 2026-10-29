import { useEffect, useState } from "react";
import "./App.css";

type Status = "New" | "Active" | "Resolved" | "Closed";
type WorkItemType = "Task" | "Bug" | "User Story";
type TaskType = "Dev" | "Prod" | "Test";

type CreateWorkItemInput = {
  title: string;
  description: string;
  work_item_type: WorkItemType;
  task_type: TaskType;
  estimated_hours: number | null;
  assigned_to: number | null;
};

type User = {
  id: number;
  display_name: string;
};

type WorkItem = {
  id: number;
  project_id: number;
  title: string;
  description: string;
  work_item_type: string;
  task_type: string;
  status: Status;
  estimated_hours: number | null;
  sort_order: number;
  assigned_to?: number | null;
  assigned_name?: string | null;
};

type WorkItemComment = {
  comment_id: number;
  work_item_id: number;
  content: string;
  created_at: string;
};

const API_URL = "http://localhost:8080";
const PROJECT_ID = 1;

const ICONS: Record<string, string> = {
  search: '<circle cx="7" cy="7" r="4.5"/><path d="M10.5 10.5L14 14"/>',
  "chevron-down": '<path d="M3.5 6l4.5 4.5L12.5 6"/>',
  "chevron-left": '<path d="M10 3L5 8l5 5"/>',
  plus: '<path d="M8 3v10M3 8h10"/>',
  star: '<path d="M8 1.8l1.9 4 4.3.5-3.2 3 .9 4.3L8 11.4l-3.9 2.2.9-4.3-3.2-3 4.3-.5z"/>',
  people:
    '<circle cx="6" cy="5.5" r="2.2"/><path d="M2 13c0-2.5 1.8-4 4-4s4 1.5 4 4M10.5 3.5a2 2 0 010 4M12 9.5c1.3.5 2 1.7 2 3.5"/>',
  "arrow-circle":
    '<circle cx="8" cy="8" r="6"/><path d="M5 8h6M8.5 5.5L11 8l-2.5 2.5"/>',
  filter: '<path d="M2 4h12M4.5 8h7M7 12h2"/>',
  gear: '<circle cx="8" cy="8" r="2.2"/><path d="M8 1.5v2M8 12.5v2M1.5 8h2M12.5 8h2M3.4 3.4l1.4 1.4M11.2 11.2l1.4 1.4M3.4 12.6l1.4-1.4M11.2 4.8l1.4-1.4"/>',
  expand: '<path d="M9.5 2H14v4.5M6.5 14H2V9.5M14 2L9 7M2 14l5-5"/>',
  info: '<circle cx="8" cy="8" r="6"/><path d="M8 7.2V11M8 5v.1"/>',
  overview:
    '<rect x="2" y="2" width="5" height="5"/><rect x="9" y="2" width="5" height="5"/><rect x="2" y="9" width="5" height="5"/><rect x="9" y="9" width="5" height="5"/>',
  workitems:
    '<rect x="3" y="2" width="10" height="12" rx="1"/><path d="M5.5 6h5M5.5 9h5"/>',
  boards:
    '<rect x="2" y="3" width="12" height="10" rx="1"/><path d="M6 3v10M10 3v10"/>',
  backlogs: '<path d="M6 4h8M6 8h8M6 12h8M2.5 4h.1M2.5 8h.1M2.5 12h.1"/>',
  sprints: '<circle cx="8" cy="8" r="6"/><path d="M8 4.5V8l2.5 1.5"/>',
  queries: '<path d="M2 3h12L9.5 8.5V13l-3-1.5V8.5z"/>',
  story: '<path d="M3 2h9a1 1 0 011 1v10H5a2 2 0 00-2 2z"/>',
  bug: '<circle cx="8" cy="9.5" r="4.5"/><circle cx="8" cy="3.8" r="2"/>',
  task: '<rect x="3" y="3" width="10" height="11" rx="1.5"/><rect x="6" y="1.5" width="4" height="3" rx="1"/>',
  save: '<path d="M2 1.5h10l2.5 2.5v10.5H1.5V1.5z"/><path d="M4 1.5v4h7v-4M4 14.5V9h8v5.5"/>',
  refresh: '<path d="M13.5 5.5A6 6 0 1 0 14 8"/><path d="M13.5 1.5v4h-4"/>',
};

function Icon({ name, className = "" }: { name: string; className?: string }) {
  return (
    <svg
      className={`icon ${className}`}
      viewBox="0 0 16 16"
      aria-hidden="true"
      dangerouslySetInnerHTML={{ __html: ICONS[name] }}
    />
  );
}

const TYPE_STYLE: Record<string, { color: string; icon: string }> = {
  Task: { color: "#00a3e0", icon: "task" },
  Bug: { color: "#d13438", icon: "bug" },
  "User Story": { color: "#8764b8", icon: "story" },
};

const TAG_CLASS: Record<string, string> = {
  Dev: "tag-dev",
  Prod: "tag-prod",
  Test: "tag-test",
};

const WIP_LIMITS: Partial<Record<Status, number>> = { Active: 5 };

function WorkItemCard({
  item,
  onStatusChange,
  onSelect,
  selectedWorkItemId,
  editedTitle,
}: {
  item: WorkItem;
  onStatusChange: (id: number, status: Status) => void;
  onSelect: (item: WorkItem) => void;
  selectedWorkItemId: number | null;
  editedTitle: string;
}) {
  const type = TYPE_STYLE[item.work_item_type] ?? TYPE_STYLE["User Story"];

  return (
    <article
      className="work-item-card"
      onClick={() => onSelect(item)}
      style={{ "--type-color": type.color } as React.CSSProperties}
      draggable
      onDragStart={(event) =>
        event.dataTransfer.setData("text/plain", String(item.id))
      }
    >
      <h3>
        <Icon
          name={item.work_item_type === "Bug" ? "bug" : type.icon}
          className={`type-icon ${
            item.work_item_type === "Bug" ? "bug-animated" : ""
          }`}
        />
        <span>{item.id === selectedWorkItemId ? editedTitle : item.title}</span>
      </h3>

      {item.description && <p className="card-desc">{item.description}</p>}

      {item.assigned_to && (
        <div className="card-assignee">
          <span className="avatar">
            {(item.assigned_name || String(item.assigned_to))[0].toUpperCase()}
          </span>
          <span className="assignee-name">
            {item.assigned_name || `User ${item.assigned_to}`}
          </span>
        </div>
      )}

      <p className="card-meta">
        Estimated hours <b>{item.estimated_hours ?? "–"}</b>
      </p>

      <div className="tags">
        <span className={`tag ${TAG_CLASS[item.task_type] ?? ""}`}>
          {item.task_type}
        </span>
      </div>

      <label className="status-control">
        <span className="sr-only">Change status for {item.title}</span>
        <select
          className="status-select"
          onClick={(event) => event.stopPropagation()}
          aria-label={`Change status for ${item.title}`}
          value={item.status}
          onChange={(event) =>
            onStatusChange(item.id, event.target.value as Status)
          }
        >
          <option value="New">New</option>
          <option value="Active">Active</option>
          <option value="Resolved">Resolved</option>
          <option value="Closed">Closed</option>
        </select>
      </label>
    </article>
  );
}

function App() {
  const [workItems, setWorkItems] = useState<WorkItem[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [usersError, setUsersError] = useState("");
  const [editedAssignedTo, setEditedAssignedTo] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [statusError, setStatusError] = useState("");
  const [selectedWorkItemId, setSelectedWorkItemId] = useState<number | null>(
    null,
  );
  const [workItemDetails, setWorkItemDetails] = useState<WorkItem | null>(null);
  const [detailsError, setDetailsError] = useState("");
  const [detailsLoading, setDetailsLoading] = useState(false);
  const [selectedWorkItem, setSelectedWorkItem] = useState<WorkItem | null>(
    null,
  );
  const [editedTitle, setEditedTitle] = useState("");
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState("");
  const [editedDescription, setEditedDescription] = useState("");
  const [editedTaskType, setEditedTaskType] = useState("Dev");
  const [editedHours, setEditedHours] = useState("");
  const [saveSuccess, setSaveSuccess] = useState("");
  const [assignedTo, setAssignedTo] = useState("");
  const [isEditing, setIsEditing] = useState(false);

  const [comments, setComments] = useState<WorkItemComment[]>([]);
  const [newComment, setNewComment] = useState("");
  const [commentsLoading, setCommentsLoading] = useState(false);
  const [commentError, setCommentError] = useState("");

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [workItemType, setWorkItemType] = useState<WorkItemType>("Task");
  const [taskType, setTaskType] = useState<TaskType>("Dev");
  const [estimatedHours, setEstimatedHours] = useState("");
  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState("");

  const statuses: Status[] = ["New", "Active", "Resolved", "Closed"];
  const [dragOver, setDragOver] = useState<Status | null>(null);

  function handleDrop(event: React.DragEvent, status: Status) {
    event.preventDefault();
    setDragOver(null);
    const id = Number(event.dataTransfer.getData("text/plain"));
    const item = workItems.find((i) => i.id === id);
    if (item && item.status !== status) void updateWorkItemStatus(id, status);
  }

  async function loadWorkItems() {
    try {
      setError("");

      const response = await fetch(
        `${API_URL}/v1/projects/${PROJECT_ID}/work-items`,
      );

      if (!response.ok) {
        throw new Error(`API request failed: ${response.status}`);
      }

      const data: WorkItem[] = await response.json();
      setWorkItems(data);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Could not load work items",
      );
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    if (!saveSuccess) return;

    const timeoutId = setTimeout(() => {
      setSaveSuccess("");
    }, 3000);

    return () => clearTimeout(timeoutId);
  }, [saveSuccess]);

  useEffect(() => {
    void loadWorkItems();
    void loadUsers();
  }, []);

  async function loadUsers() {
    try {
      setUsersError("");
      const response = await fetch(`${API_URL}/v1/users`);
      if (!response.ok) {
        throw new Error(`Could not load users (HTTP ${response.status})`);
      }
      const data: User[] = await response.json();
      setUsers(data);
    } catch (err) {
      setUsersError(
        err instanceof Error ? err.message : "Could not load users",
      );
    }
  }

  useEffect(() => {
    if (selectedWorkItemId === null) {
      setComments([]);
      return;
    }

    const fetchComments = async () => {
      setCommentsLoading(true);
      setCommentError("");

      try {
        const response = await fetch(
          `${API_URL}/v1/work-items/${selectedWorkItemId}/comments`,
        );

        if (!response.ok) {
          throw new Error("Failed to load comments");
        }

        const data: WorkItemComment[] = await response.json();
        setComments(data);
      } catch (error) {
        setCommentError(
          error instanceof Error ? error.message : "Something went wrong",
        );
      } finally {
        setCommentsLoading(false);
      }
    };

    fetchComments();
  }, [selectedWorkItemId]);

  async function handleSubmitComment(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (selectedWorkItemId === null || !newComment.trim()) {
      return;
    }

    setCommentError("");

    try {
      const response = await fetch(
        `${API_URL}/v1/work-items/${selectedWorkItemId}/comments`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            content: newComment.trim(),
          }),
        },
      );

      if (!response.ok) {
        throw new Error((await response.text()) || "Failed to post comment");
      }

      const createdComment: WorkItemComment = await response.json();

      setComments((current) => [...current, createdComment]);
      setNewComment("");
    } catch (err) {
      setCommentError(
        err instanceof Error ? err.message : "Could not post comment",
      );
    }
  }

  // Fetch full details whenever the selected card ID changes.
  useEffect(() => {
    if (selectedWorkItemId === null) {
      setWorkItemDetails(null);
      setDetailsError("");
      setDetailsLoading(false);
      return;
    }

    const controller = new AbortController();
    setWorkItemDetails(null);
    setDetailsError("");
    setDetailsLoading(true);
    setSaveSuccess("");
    setSaveError("");

    async function loadDetails() {
      try {
        const response = await fetch(
          `${API_URL}/v1/work-items/${selectedWorkItemId}`,
          { signal: controller.signal },
        );
        if (!response.ok) {
          throw new Error(`Could not load details (HTTP ${response.status})`);
        }
        const data: WorkItem = await response.json();
        if (!controller.signal.aborted) {
          setWorkItemDetails(data);
          setEditedTitle(data.title);
          setEditedDescription(data.description ?? "");
          setEditedTaskType(data.task_type ?? "Dev");
          setEditedAssignedTo(
            data.assigned_to == null ? "" : String(data.assigned_to),
          );
          setEditedHours(
            data.estimated_hours == null ? "" : String(data.estimated_hours),
          );
        }
      } catch (err) {
        if (!controller.signal.aborted) {
          setDetailsError(
            err instanceof Error ? err.message : "Could not load details",
          );
        }
      } finally {
        if (!controller.signal.aborted) setDetailsLoading(false);
      }
    }

    void loadDetails();
    return () => controller.abort();
  }, [selectedWorkItemId]);

  async function updateWorkItemStatus(id: number, newStatus: Status) {
    setStatusError("");

    try {
      const response = await fetch(`${API_URL}/v1/work-items/${id}/status`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status: newStatus }),
      });

      if (!response.ok) {
        throw new Error(
          (await response.text()) || `Request failed: ${response.status}`,
        );
      }

      const updatedItem: WorkItem = await response.json();

      setWorkItems((currentItems) =>
        currentItems.map((item) =>
          item.id === updatedItem.id ? updatedItem : item,
        ),
      );
    } catch (err) {
      setStatusError(
        err instanceof Error ? err.message : "Could not update status",
      );
    }
  }

  async function saveWorkItem() {
    if (selectedWorkItemId === null || !workItemDetails) {
      return;
    }

    if (
      editedHours.trim() !== "" &&
      (!Number.isFinite(Number(editedHours)) || Number(editedHours) < 0)
    ) {
      setSaveError("Estimated hours must be a valid non-negative number.");
      return;
    }

    setSaving(true);
    setSaveError("");
    setSaveSuccess("");

    try {
      const response = await fetch(
        `${API_URL}/v1/work-items/${selectedWorkItemId}`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            ...workItemDetails,
            title: editedTitle,
            description: editedDescription,
            task_type: editedTaskType,
            assigned_to:
              editedAssignedTo === "" ? null : Number(editedAssignedTo),
            estimated_hours:
              editedHours.trim() === "" ? null : Number(editedHours),
          }),
        },
      );

      if (!response.ok) {
        throw new Error(`Save failed (HTTP ${response.status})`);
      }

      const updatedItem: WorkItem = await response.json();

      setWorkItemDetails(updatedItem);
      setSaveSuccess("All changes saved");

      setEditedDescription(updatedItem.description ?? "");
      setEditedTaskType(updatedItem.task_type || "Dev");
      setEditedHours(
        updatedItem.estimated_hours == null
          ? ""
          : String(updatedItem.estimated_hours),
      );
      setIsEditing(false);

      setWorkItems((items) =>
        items.map((item) => (item.id === updatedItem.id ? updatedItem : item)),
      );

      setEditedTitle(updatedItem.title);
      setEditedAssignedTo(
        updatedItem.assigned_to == null ? "" : String(updatedItem.assigned_to),
      );
    } catch (err) {
      setSaveError(
        err instanceof Error ? err.message : "Could not save work item",
      );
    } finally {
      setSaving(false);
    }
  }

  async function refreshWorkItem() {
    if (selectedWorkItemId === null) return;

    if (isEditing && hasUnsavedChanges) {
      const confirmed = window.confirm(
        "Discard unsaved changes and reload the work item?",
      );
      if (!confirmed) return;
    }

    setDetailsLoading(true);
    setDetailsError("");

    try {
      const response = await fetch(
        `${API_URL}/v1/work-items/${selectedWorkItemId}`,
      );

      if (!response.ok) {
        throw new Error(`Refresh failed (HTTP ${response.status})`);
      }

      const data: WorkItem = await response.json();

      setWorkItemDetails(data);
      setEditedTitle(data.title);
      setEditedDescription(data.description ?? "");
      setEditedTaskType(data.task_type ?? "Dev");
      setEditedAssignedTo(
        data.assigned_to == null ? "" : String(data.assigned_to),
      );
      setEditedHours(
        data.estimated_hours == null ? "" : String(data.estimated_hours),
      );

      setWorkItems((items) =>
        items.map((item) => (item.id === data.id ? data : item)),
      );

      setIsEditing(false);
      setSaveError("");
      setSaveSuccess("");
    } catch (err) {
      setDetailsError(
        err instanceof Error ? err.message : "Could not refresh work item",
      );
    } finally {
      setDetailsLoading(false);
    }
  }

  async function createWorkItem(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setCreateError("");

    if (!title.trim()) {
      setCreateError("Please enter a title.");
      return;
    }

    const hours = estimatedHours.trim() === "" ? null : Number(estimatedHours);

    if (hours !== null && (!Number.isFinite(hours) || hours < 0)) {
      setCreateError("Estimated hours must be a valid non-negative number.");
      return;
    }

    const input: CreateWorkItemInput = {
      title: title.trim(),
      description: description.trim(),
      work_item_type: workItemType,
      task_type: taskType,
      estimated_hours: hours,
      assigned_to: assignedTo.trim() === "" ? null : Number(assignedTo),
    };

    setCreating(true);

    try {
      const response = await fetch(
        `${API_URL}/v1/projects/${PROJECT_ID}/work-items`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(input),
        },
      );

      if (!response.ok) {
        throw new Error(
          (await response.text()) || `Request failed: ${response.status}`,
        );
      }

      const createdItem: WorkItem = await response.json();

      setWorkItems((currentItems) => [...currentItems, createdItem]);
      setTitle("");
      setDescription("");
      setWorkItemType("Task");
      setTaskType("Dev");
      setEstimatedHours("");
      setShowCreateForm(false);
    } catch (err) {
      setCreateError(
        err instanceof Error ? err.message : "Could not create work item",
      );
    } finally {
      setCreating(false);
    }
  }

  const hasUnsavedChanges =
    workItemDetails !== null &&
    (editedTitle !== workItemDetails.title ||
      editedDescription !== (workItemDetails.description ?? "") ||
      editedTaskType !== (workItemDetails.task_type ?? "Dev") ||
      editedAssignedTo !==
        (workItemDetails.assigned_to == null
          ? ""
          : String(workItemDetails.assigned_to)) ||
      editedHours !==
        (workItemDetails.estimated_hours == null
          ? ""
          : String(workItemDetails.estimated_hours)));

  function handleWorkItemSelect(item: WorkItem) {
    if (item.id === selectedWorkItemId) return;

    if (isEditing && hasUnsavedChanges) {
      const confirmed = window.confirm(
        "You have unsaved changes. Discard them and switch work items?",
      );

      if (!confirmed) return;
    }

    setIsEditing(false);
    setSelectedWorkItemId(item.id);
  }

  return (
    <>
      <header className="top-bar">
        <div className="brand">
          <Icon name="boards" /> Workboard
        </div>
        <span className="org">My Organization</span>
        <div className="breadcrumbs">
          <span>/</span>
          <span>My Project</span>
          <span>/</span>
          <span>Boards</span>
          <span>/</span>
          <span>Boards</span>
        </div>
        <label className="top-search">
          <Icon name="search" />
          <input
            type="search"
            placeholder="Search"
            aria-label="Search work items"
            onChange={(event) => {
              const query = event.target.value.toLowerCase();
              // Search filtering is added in a later step.
              void query;
            }}
          />
        </label>
        <Icon name="backlogs" />
      </header>

      <div className="app-layout">
        <aside className="sidebar">
          <div className="project-heading">
            <span className="avatar-sq">MP</span>
            <span>My Project</span>
            <span className="push">
              <Icon name="plus" />
            </span>
          </div>

          <nav className="sidebar-nav" aria-label="Project navigation">
            <a className="sidebar-link" href="#overview">
              <Icon name="overview" />
              Overview
            </a>
            <div className="sidebar-group">
              <a className="sidebar-link group-title" href="#boards">
                <span className="group-icon">
                  <Icon name="boards" />
                </span>
                Boards
              </a>
              <a className="sidebar-link" href="#work-items">
                <Icon name="workitems" />
                Work items
              </a>
              <a className="sidebar-link active" href="#board">
                <Icon name="boards" />
                Boards
              </a>
              <a className="sidebar-link" href="#backlogs">
                <Icon name="backlogs" />
                Backlogs
              </a>
              <a className="sidebar-link" href="#sprints">
                <Icon name="sprints" />
                Sprints
              </a>
              <a className="sidebar-link" href="#queries">
                <Icon name="queries" />
                Queries
              </a>
            </div>
          </nav>
        </aside>

        <main className="board-workspace" id="board">
          <div className="team-heading">
            <Icon name="boards" className="team-icon" />
            <h1>My Project Team</h1>
            <Icon name="chevron-down" />
            <Icon name="star" className="star" />
            <Icon name="people" />
            <button type="button" className="view-backlog">
              <Icon name="arrow-circle" />
              View as backlog
            </button>
          </div>

          <div className="board-toolbar">
            <button className="toolbar-tab active" type="button">
              Board
            </button>
            <button className="toolbar-tab" type="button">
              Analytics
            </button>
            <div className="toolbar-actions">
              <span className="stories">
                <Icon name="boards" />
                Stories
                <Icon name="chevron-down" />
              </span>
              <Icon name="filter" />
              <Icon name="gear" />
              <Icon name="expand" />
            </div>
          </div>

          {statusError && <p role="alert">{statusError}</p>}
          {error && <p role="alert">Could not load work items: {error}</p>}

          {showCreateForm && (
            <form className="create-work-item-form" onSubmit={createWorkItem}>
              <div className="form-heading">
                <h2>Create Work Item</h2>
                <button
                  type="button"
                  className="close-form-button"
                  onClick={() => {
                    setShowCreateForm(false);
                    setCreateError("");
                  }}
                  aria-label="Close form"
                >
                  ×
                </button>
              </div>

              <label htmlFor="title">Title *</label>
              <input
                id="title"
                type="text"
                value={title}
                onChange={(event) => setTitle(event.target.value)}
                placeholder="Enter work item title"
                required
              />

              <label htmlFor="description">Description</label>
              <textarea
                id="description"
                value={description}
                onChange={(event) => setDescription(event.target.value)}
                placeholder="Describe the work item"
              />

              <label htmlFor="workItemType">Work Item Type</label>
              <select
                id="workItemType"
                value={workItemType}
                onChange={(event) =>
                  setWorkItemType(event.target.value as WorkItemType)
                }
              >
                <option value="Task">Task</option>
                <option value="Bug">Bug</option>
                <option value="User Story">User Story</option>
              </select>

              <label htmlFor="taskType">Task Type</label>
              <select
                id="taskType"
                value={taskType}
                onChange={(event) =>
                  setTaskType(event.target.value as TaskType)
                }
              >
                <option value="Dev">Dev</option>
                <option value="Prod">Prod</option>
                <option value="Test">Test</option>
              </select>

              <label htmlFor="create-assigned-to">Assign To</label>

              <select
                id="create-assigned-to"
                value={assignedTo}
                onChange={(event) => setAssignedTo(event.target.value)}
              >
                <option value="">Unassigned</option>

                {users.map((user) => (
                  <option key={user.id} value={String(user.id)}>
                    {user.display_name}
                  </option>
                ))}
              </select>

              {usersError && <p role="alert">{usersError}</p>}

              <label htmlFor="estimatedHours">Estimated Hours</label>
              <input
                id="estimatedHours"
                type="number"
                min="0"
                step="0.5"
                value={estimatedHours}
                onChange={(event) => setEstimatedHours(event.target.value)}
                placeholder="e.g. 4"
              />

              {createError && <p role="alert">{createError}</p>}

              <button type="submit" disabled={creating}>
                {creating ? "Creating..." : "Create Work Item"}
              </button>
            </form>
          )}

          {loading && <p>Loading work items...</p>}

          {!loading && !error && (
            <div className="board">
              {statuses.map((columnStatus) => {
                const items = workItems.filter(
                  (item) => item.status === columnStatus,
                );
                const limit = WIP_LIMITS[columnStatus];

                return (
                  <section
                    className={`board-column${dragOver === columnStatus ? " drag-over" : ""}`}
                    key={columnStatus}
                    onDragOver={(event) => {
                      event.preventDefault();
                      setDragOver(columnStatus);
                    }}
                    onDragLeave={() => setDragOver(null)}
                    onDrop={(event) => handleDrop(event, columnStatus)}
                  >
                    <div className="column-header">
                      <h2>
                        {columnStatus}
                        {limit && <Icon name="info" />}
                      </h2>
                      {limit ? (
                        <span
                          className={`column-count${items.length > limit ? " over" : ""}`}
                        >
                          {items.length}
                          <small>/{limit}</small>
                        </span>
                      ) : (
                        <Icon name="chevron-left" />
                      )}
                    </div>

                    <div className="column-body">
                      {columnStatus === "New" && (
                        <div className="column-actions">
                          <button
                            type="button"
                            className="new-item-button"
                            onClick={() => {
                              setCreateError("");
                              setShowCreateForm(true);
                            }}
                          >
                            <Icon name="plus" />
                            New item
                          </button>
                          <button
                            type="button"
                            className="icon-button"
                            aria-label="Search column"
                          >
                            <Icon name="search" />
                          </button>
                        </div>
                      )}

                      {items.map((item) => (
                        <WorkItemCard
                          key={item.id}
                          item={item}
                          onStatusChange={updateWorkItemStatus}
                          onSelect={handleWorkItemSelect}
                          selectedWorkItemId={selectedWorkItemId}
                          editedTitle={editedTitle}
                        />
                      ))}
                    </div>
                  </section>
                );
              })}
            </div>
          )}
          {selectedWorkItemId !== null && (
            <section
              className="work-item-details-panel"
              aria-label="Work item details"
            >
              <header className="details-panel-header">
                <div className="details-header-content">
                  <div className="details-work-item-type">
                    <Icon
                      name={
                        workItemDetails?.work_item_type === "Bug"
                          ? "bug"
                          : workItemDetails?.work_item_type === "Task"
                            ? "task"
                            : "story"
                      }
                    />
                    <span>
                      {workItemDetails?.work_item_type.toUpperCase()}{" "}
                      {workItemDetails?.id}
                    </span>
                  </div>

                  <h2 className="details-work-item-title">
                    <span>{workItemDetails?.id}</span>
                    {isEditing ? (
                      <input
                        value={editedTitle}
                        onChange={(event) => {
                          setEditedTitle(event.target.value);
                          setSaveSuccess("");
                        }}
                        aria-label="Work item title"
                      />
                    ) : (
                      <span>{workItemDetails?.title}</span>
                    )}
                  </h2>
                </div>

                <div className="work-item-header-actions">
                  {!isEditing && (
                    <button
                      type="button"
                      className="header-action-button"
                      onClick={() => setIsEditing(true)}
                      disabled={detailsLoading || !workItemDetails}
                    >
                      Edit
                    </button>
                  )}

                  {isEditing && (
                    <>
                      <button
                        type="button"
                        onClick={saveWorkItem}
                        disabled={
                          saving || !editedTitle.trim() || !hasUnsavedChanges
                        }
                      >
                        {saving ? "Saving..." : "Save"}
                      </button>

                      <button
                        type="button"
                        onClick={() => {
                          if (!workItemDetails) return;

                          setEditedTitle(workItemDetails.title);
                          setEditedDescription(
                            workItemDetails.description ?? "",
                          );
                          setEditedTaskType(workItemDetails.task_type || "Dev");
                          setEditedAssignedTo(
                            workItemDetails.assigned_to == null
                              ? ""
                              : String(workItemDetails.assigned_to),
                          );
                          setEditedHours(
                            workItemDetails.estimated_hours == null
                              ? ""
                              : String(workItemDetails.estimated_hours),
                          );

                          setSaveError("");
                          setSaveSuccess("");
                          setIsEditing(false);
                        }}
                      >
                        Cancel
                      </button>
                    </>
                  )}

                  <button
                    type="button"
                    onClick={() => {
                      if (hasUnsavedChanges) {
                        const confirmed = window.confirm(
                          "You have unsaved changes. Close without saving?",
                        );
                        if (!confirmed) return;
                      }

                      setSelectedWorkItemId(null);
                      setIsEditing(false);
                      setSaveError("");
                      setSaveSuccess("");
                    }}
                  >
                    Close
                  </button>

                  <button
                    type="button"
                    className="header-action-button refresh-action"
                    onClick={refreshWorkItem}
                    disabled={detailsLoading || saving || !workItemDetails}
                    title="Refresh work item"
                    aria-label="Refresh work item"
                  >
                    <Icon name="refresh" />
                  </button>
                </div>
              </header>

              {workItemDetails && (
                <div className="work-item-identity-row">
                  <div className="work-item-assignee">
                    <span className="avatar">
                      {workItemDetails.assigned_name
                        ? workItemDetails.assigned_name[0].toUpperCase()
                        : workItemDetails.assigned_to != null
                          ? String(workItemDetails.assigned_to)[0]
                          : "?"}
                    </span>

                    <span>
                      {workItemDetails.assigned_name ||
                        (workItemDetails.assigned_to != null
                          ? `User ${workItemDetails.assigned_to}`
                          : "Unassigned")}
                    </span>
                  </div>

                  <button
                    type="button"
                    className="work-item-comment-link"
                    onClick={() => {
                      document
                        .querySelector(".discussion-section")
                        ?.scrollIntoView({
                          behavior: "smooth",
                          block: "start",
                        });
                    }}
                  >
                    <Icon name="overview" />
                    {comments.length}{" "}
                    {comments.length === 1 ? "comment" : "comments"}
                  </button>

                  <div className="work-item-tags">
                    <span
                      className={`work-item-tag ${TAG_CLASS[workItemDetails.task_type] ?? ""}`}
                    >
                      {workItemDetails.task_type}
                    </span>
                  </div>

                  <button type="button" className="work-item-follow-button">
                    Follow
                  </button>
                </div>
              )}

              {detailsLoading && <p>Loading details...</p>}
              {detailsError && <p role="alert">{detailsError}</p>}
              {workItemDetails && (
                <div>
                  {saveError && <p role="alert">{saveError}</p>}

                  {saveSuccess && (
                    <p role="status" style={{ color: "green" }}>
                      {saveSuccess}
                    </p>
                  )}

                  <div className="work-item-body-grid">
                    <div className="work-item-main-column">
                      <section className="work-item-content-section">
                        <h3>Description</h3>

                        {isEditing ? (
                          <textarea
                            id="work-item-description"
                            value={editedDescription}
                            onChange={(event) => {
                              setEditedDescription(event.target.value);
                              setSaveSuccess("");
                            }}
                            rows={5}
                          />
                        ) : (
                          <p>
                            {workItemDetails.description ||
                              "Click to add Description"}
                          </p>
                        )}
                      </section>

                      <div className="discussion-section">
                        <div className="discussion-header">
                          <h3>Discussion</h3>
                          <span className="discussion-count">
                            {comments.length}{" "}
                            {comments.length === 1 ? "comment" : "comments"}
                          </span>
                        </div>

                        <form
                          className="discussion-form"
                          onSubmit={handleSubmitComment}
                        >
                          <div className="discussion-avatar current-user-avatar">
                            Y
                          </div>

                          <div className="discussion-composer">
                            <textarea
                              value={newComment}
                              onChange={(event) =>
                                setNewComment(event.target.value)
                              }
                              placeholder="Add a comment. Use @ to mention a person."
                              aria-label="Write a comment"
                              rows={3}
                            />

                            <div className="discussion-composer-actions">
                              <button
                                type="submit"
                                disabled={!newComment.trim()}
                              >
                                Comment
                              </button>
                            </div>
                          </div>
                        </form>

                        {commentError && (
                          <p className="discussion-error" role="alert">
                            {commentError}
                          </p>
                        )}

                        {commentsLoading && (
                          <p className="discussion-muted">
                            Loading comments...
                          </p>
                        )}

                        {!commentsLoading &&
                          !commentError &&
                          comments.length === 0 && (
                            <p className="discussion-muted">
                              No comments yet. Start the discussion!
                            </p>
                          )}

                        <div className="discussion-list">
                          {comments.map((comment) => (
                            <article
                              key={comment.comment_id}
                              className="discussion-comment"
                            >
                              <div className="discussion-avatar comment-avatar">
                                C
                              </div>

                              <div className="discussion-comment-card">
                                <div className="discussion-comment-meta">
                                  <strong>Comment</strong>
                                  <span>
                                    {new Date(
                                      comment.created_at,
                                    ).toLocaleString()}
                                  </span>
                                </div>
                                <p className="discussion-comment-content">
                                  {comment.content}
                                </p>
                              </div>
                            </article>
                          ))}
                        </div>
                      </div>
                    </div>

                    <aside className="work-item-sidebar-column">
                      <section className="work-item-content-section">
                        <h3>Planning</h3>

                        <div className="work-item-field">
                          <label>Status</label>
                          <p className="work-item-field-value">
                            <span
                              className={`status-dot status-${workItemDetails.status.toLowerCase()}`}
                            />
                            {workItemDetails.status}
                          </p>
                        </div>

                        <div className="work-item-field">
                          <label htmlFor="work-item-estimated-hours">
                            Estimated Hours
                          </label>
                          {isEditing ? (
                            <input
                              id="work-item-estimated-hours"
                              type="number"
                              min="0"
                              step="0.5"
                              value={editedHours}
                              onChange={(event) => {
                                setEditedHours(event.target.value);
                                setSaveSuccess("");
                                setSaveError("");
                              }}
                            />
                          ) : (
                            <p className="work-item-field-value">
                              {workItemDetails.estimated_hours == null
                                ? "Not set"
                                : `${workItemDetails.estimated_hours} hours`}
                            </p>
                          )}
                        </div>
                      </section>

                      <section className="work-item-content-section">
                        <h3>Classification</h3>

                        <div className="work-item-field">
                          <label htmlFor="work-item-task-type">Task Type</label>
                          {isEditing ? (
                            <select
                              id="work-item-task-type"
                              value={editedTaskType}
                              onChange={(event) => {
                                setEditedTaskType(event.target.value);
                                setSaveSuccess("");
                              }}
                            >
                              <option value="Dev">Dev</option>
                              <option value="Prod">Prod</option>
                              <option value="Test">Test</option>
                            </select>
                          ) : (
                            <p className="work-item-field-value">
                              {workItemDetails.task_type || "Not set"}
                            </p>
                          )}
                        </div>

                        <div className="work-item-field">
                          <label htmlFor="work-item-assignee">
                            Assigned To
                          </label>
                          {isEditing ? (
                            <select
                              id="work-item-assignee"
                              value={editedAssignedTo}
                              onChange={(event) => {
                                setEditedAssignedTo(event.target.value);
                                setSaveSuccess("");
                                setSaveError("");
                              }}
                            >
                              <option value="">Unassigned</option>
                              {users.map((user) => (
                                <option key={user.id} value={String(user.id)}>
                                  {user.display_name}
                                </option>
                              ))}
                            </select>
                          ) : (
                            <p className="work-item-field-value">
                              {workItemDetails.assigned_name || "Unassigned"}
                            </p>
                          )}
                          {usersError && <p role="alert">{usersError}</p>}
                        </div>
                      </section>
                    </aside>
                  </div>
                </div>
              )}
            </section>
          )}
        </main>
      </div>
    </>
  );
}

export default App;
