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
  assigned_to?: string | null;
};

const API_URL = "http://localhost:8080";
const PROJECT_ID = 1;

const ICONS: Record<string, string> = {
  search: '<circle cx="7" cy="7" r="4.5"/><path d="M10.5 10.5L14 14"/>',
  "chevron-down": '<path d="M3.5 6l4.5 4.5L12.5 6"/>',
  "chevron-left": '<path d="M10 3L5 8l5 5"/>',
  plus: '<path d="M8 3v10M3 8h10"/>',
  star: '<path d="M8 1.8l1.9 4 4.3.5-3.2 3 .9 4.3L8 11.4l-3.9 2.2.9-4.3-3.2-3 4.3-.5z"/>',
  people: '<circle cx="6" cy="5.5" r="2.2"/><path d="M2 13c0-2.5 1.8-4 4-4s4 1.5 4 4M10.5 3.5a2 2 0 010 4M12 9.5c1.3.5 2 1.7 2 3.5"/>',
  "arrow-circle": '<circle cx="8" cy="8" r="6"/><path d="M5 8h6M8.5 5.5L11 8l-2.5 2.5"/>',
  filter: '<path d="M2 4h12M4.5 8h7M7 12h2"/>',
  gear: '<circle cx="8" cy="8" r="2.2"/><path d="M8 1.5v2M8 12.5v2M1.5 8h2M12.5 8h2M3.4 3.4l1.4 1.4M11.2 11.2l1.4 1.4M3.4 12.6l1.4-1.4M11.2 4.8l1.4-1.4"/>',
  expand: '<path d="M9.5 2H14v4.5M6.5 14H2V9.5M14 2L9 7M2 14l5-5"/>',
  info: '<circle cx="8" cy="8" r="6"/><path d="M8 7.2V11M8 5v.1"/>',
  overview: '<rect x="2" y="2" width="5" height="5"/><rect x="9" y="2" width="5" height="5"/><rect x="2" y="9" width="5" height="5"/><rect x="9" y="9" width="5" height="5"/>',
  workitems: '<rect x="3" y="2" width="10" height="12" rx="1"/><path d="M5.5 6h5M5.5 9h5"/>',
  boards: '<rect x="2" y="3" width="12" height="10" rx="1"/><path d="M6 3v10M10 3v10"/>',
  backlogs: '<path d="M6 4h8M6 8h8M6 12h8M2.5 4h.1M2.5 8h.1M2.5 12h.1"/>',
  sprints: '<circle cx="8" cy="8" r="6"/><path d="M8 4.5V8l2.5 1.5"/>',
  queries: '<path d="M2 3h12L9.5 8.5V13l-3-1.5V8.5z"/>',
  story: '<path d="M3 2h9a1 1 0 011 1v10H5a2 2 0 00-2 2z"/>',
  bug: '<circle cx="8" cy="9.5" r="4.5"/><circle cx="8" cy="3.8" r="2"/>',
  task: '<rect x="3" y="3" width="10" height="11" rx="1.5"/><rect x="6" y="1.5" width="4" height="3" rx="1"/>',
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
}: {
  item: WorkItem;
  onStatusChange: (id: number, status: Status) => void;
}) {
  const type = TYPE_STYLE[item.work_item_type] ?? TYPE_STYLE["User Story"];

  return (
    <article
      className="work-item-card"
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
        <span>{item.title}</span>
      </h3>

      {item.description && (
        <p className="card-desc">{item.description}</p>
      )}

      {item.assigned_to && (
        <div className="card-assignee">
          <span className="avatar">
            {item.assigned_to[0].toUpperCase()}
          </span>
          <span className="assignee-name">{item.assigned_to}</span>
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
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [statusError, setStatusError] = useState("");

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [workItemType, setWorkItemType] =
    useState<WorkItemType>("Task");
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
        `${API_URL}/v1/projects/${PROJECT_ID}/work-items`
      );

      if (!response.ok) {
        throw new Error(`API request failed: ${response.status}`);
      }

      const data: WorkItem[] = await response.json();
      setWorkItems(data);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Could not load work items"
      );
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadWorkItems();
  }, []);

  async function updateWorkItemStatus(id: number, newStatus: Status) {
    setStatusError("");

    try {
      const response = await fetch(
        `${API_URL}/v1/work-items/${id}/status`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ status: newStatus }),
        }
      );

      if (!response.ok) {
        throw new Error(
          (await response.text()) || `Request failed: ${response.status}`
        );
      }

      const updatedItem: WorkItem = await response.json();

      setWorkItems((currentItems) =>
        currentItems.map((item) =>
          item.id === updatedItem.id ? updatedItem : item
        )
      );
    } catch (err) {
      setStatusError(
        err instanceof Error ? err.message : "Could not update status"
      );
    }
  }

  async function createWorkItem(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setCreateError("");

    if (!title.trim()) {
      setCreateError("Please enter a title.");
      return;
    }

    const hours =
      estimatedHours.trim() === "" ? null : Number(estimatedHours);

    if (
      hours !== null &&
      (!Number.isFinite(hours) || hours < 0)
    ) {
      setCreateError("Estimated hours must be a valid non-negative number.");
      return;
    }

    const input: CreateWorkItemInput = {
      title: title.trim(),
      description: description.trim(),
      work_item_type: workItemType,
      task_type: taskType,
      estimated_hours: hours,
    };

    setCreating(true);

    try {
      const response = await fetch(
        `${API_URL}/v1/projects/${PROJECT_ID}/work-items`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(input),
        }
      );

      if (!response.ok) {
        throw new Error(
          (await response.text()) || `Request failed: ${response.status}`
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
        err instanceof Error ? err.message : "Could not create work item"
      );
    } finally {
      setCreating(false);
    }
  }

  return (
    <>
      <header className="top-bar">
        <div className="brand">
          <Icon name="boards" /> Workboard
        </div>
        <span className="org">My Organization</span>
        <div className="breadcrumbs">
          <span>/</span><span>My Project</span><span>/</span>
          <span>Boards</span><span>/</span><span>Boards</span>
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
            <span className="push"><Icon name="plus" /></span>
          </div>

          <nav className="sidebar-nav" aria-label="Project navigation">
            <a className="sidebar-link" href="#overview"><Icon name="overview" />Overview</a>
            <div className="sidebar-group">
              <a className="sidebar-link group-title" href="#boards">
                <span className="group-icon"><Icon name="boards" /></span>Boards
              </a>
              <a className="sidebar-link" href="#work-items"><Icon name="workitems" />Work items</a>
              <a className="sidebar-link active" href="#board"><Icon name="boards" />Boards</a>
              <a className="sidebar-link" href="#backlogs"><Icon name="backlogs" />Backlogs</a>
              <a className="sidebar-link" href="#sprints"><Icon name="sprints" />Sprints</a>
              <a className="sidebar-link" href="#queries"><Icon name="queries" />Queries</a>
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
              <Icon name="arrow-circle" />View as backlog
            </button>
          </div>

          <div className="board-toolbar">
            <button className="toolbar-tab active" type="button">Board</button>
            <button className="toolbar-tab" type="button">Analytics</button>
            <div className="toolbar-actions">
              <span className="stories">
                <Icon name="boards" />Stories<Icon name="chevron-down" />
              </span>
              <Icon name="filter" />
              <Icon name="gear" />
              <Icon name="expand" />
            </div>
          </div>

          {statusError && <p role="alert">{statusError}</p>}
          {error && <p role="alert">Could not load work items: {error}</p>}

          {showCreateForm && (
            <form
              className="create-work-item-form"
              onSubmit={createWorkItem}
            >
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

              <label htmlFor="estimatedHours">Estimated Hours</label>
              <input
                id="estimatedHours"
                type="number"
                min="0"
                step="0.5"
                value={estimatedHours}
                onChange={(event) =>
                  setEstimatedHours(event.target.value)
                }
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
                  (item) => item.status === columnStatus
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
                            <Icon name="plus" />New item
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
                        />
                      ))}
                    </div>
                  </section>
                );
              })}
            </div>
          )}
        </main>
      </div>
    </>
  );
}

export default App;
