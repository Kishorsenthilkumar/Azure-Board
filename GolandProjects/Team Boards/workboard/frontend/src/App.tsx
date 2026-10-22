
import { useEffect, useState } from "react";
import "./App.css";

type Status = "New" | "Active" | "Resolved" | "Closed";

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
};

const API_URL = "http://localhost:8080";

function WorkItemCard({
  item,
  onStatusChange,
}: {
  item: WorkItem;
  onStatusChange: (id: number, status: Status) => void;
}) {
  return (
    <div className="work-item-card">
      <h3>{item.title}</h3>
      <p>Status: {item.status}</p>
      <p>Type: {item.work_item_type} · {item.task_type}</p>
      <p>Estimated hours: {item.estimated_hours ?? "Not set"}</p>

      <button onClick={() => onStatusChange(item.id, "New")}>
        Set New
      </button>
      <button onClick={() => onStatusChange(item.id, "Active")}>
        Set Active
      </button>
      <button onClick={() => onStatusChange(item.id, "Resolved")}>
        Set Resolved
      </button>
      <button onClick={() => onStatusChange(item.id, "Closed")}>
        Set Closed
      </button>
    </div>
  );
}

function App() {
  const [workItems, setWorkItems] = useState<WorkItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [statusError, setStatusError] = useState("");

  useEffect(() => {
    async function loadWorkItems() {
      try {
        const response = await fetch(
          `${API_URL}/v1/projects/1/work-items`
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

    loadWorkItems();
  }, []);

  
async function updateWorkItemStatus(id: number, newStatus: Status) {
  setStatusError("");

  try {
    const response = await fetch(
      `${API_URL}/v1/work-items/${id}/status`,
      {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ status: newStatus }),
      }
    );

    if (!response.ok) {
      const message = await response.text();
      throw new Error(message || `Request failed: ${response.status}`);
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


  const statuses: Status[] = ["New", "Active", "Resolved", "Closed"];

  return (
    <div>
      <h1>Workboard</h1>
      {statusError && <p role="alert">{statusError}</p>}
      {loading && <p>Loading work items...</p>}
      {error && <p>Could not load work items: {error}</p>}

      {!loading && !error && (
        <div className="board">
          {statuses.map((columnStatus) => (
            <section className="board-column" key={columnStatus}>
              <h2>{columnStatus}</h2>

              {workItems
                .filter((item) => item.status === columnStatus)
                .map((item) => (
                  <WorkItemCard
                    key={item.id}
                    item={item}
                    onStatusChange={updateWorkItemStatus}
                  />
                ))}
            </section>
          ))}
        </div>
      )}
    </div>
  );
}

export default App;
