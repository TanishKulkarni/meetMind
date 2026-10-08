"use client";

import { useEffect, useState } from "react";
import Link from "next/link";

const API_URL = "http://localhost:8080";

export default function TasksPage() {
  const [tasks, setTasks] = useState([]);
  const [filter, setFilter] = useState("all");
  const [loading, setLoading] = useState(true);

  const fetchTasks = async () => {
    try {
      const response = await fetch(`${API_URL}/api/tasks`);

      if (!response.ok) {
        throw new Error("Failed to fetch tasks");
      }

      const data = await response.json();
      setTasks(data.tasks || []);
    } catch (error) {
      console.error("Error fetching tasks:", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTasks();
  }, []);

  const updateTaskStatus = async (taskId, status) => {
    try {
      const response = await fetch(
        `${API_URL}/api/tasks/${taskId}/status`,
        {
          method: "PATCH",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ status }),
        }
      );

      if (!response.ok) {
        throw new Error("Failed to update task");
      }

      setTasks((currentTasks) =>
        currentTasks.map((task) =>
          task.id === taskId
            ? { ...task, status }
            : task
        )
      );
    } catch (error) {
      console.error("Error updating task:", error);
    }
  };

  const filteredTasks = tasks.filter((task) => {
    if (filter === "pending") {
      return task.status === "pending";
    }

    if (filter === "completed") {
      return task.status === "completed";
    }

    return true;
  });

  const pendingCount = tasks.filter(
    (task) => task.status === "pending"
  ).length;

  const completedCount = tasks.filter(
    (task) => task.status === "completed"
  ).length;

  return (
    <div className="min-h-screen bg-gray-100">
      <div className="flex min-h-screen">
        {/* Sidebar */}
        <aside className="hidden w-64 bg-white border-r border-gray-200 md:block">
          <div className="p-6">
            <h1 className="text-2xl font-bold text-gray-900">
              Meet AI
            </h1>
            <p className="mt-1 text-sm text-gray-500">
              Meeting Intelligence
            </p>
          </div>

          <nav className="px-4 space-y-2">
            <Link
              href="/"
              className="flex items-center px-4 py-3 text-gray-600 rounded-lg hover:bg-gray-100"
            >
              Dashboard
            </Link>

            <Link
              href="/meetings"
              className="flex items-center px-4 py-3 text-gray-600 rounded-lg hover:bg-gray-100"
            >
              Meetings
            </Link>

            <Link
              href="/tasks"
              className="flex items-center px-4 py-3 font-medium text-blue-600 bg-blue-50 rounded-lg"
            >
              Tasks
            </Link>
          </nav>
        </aside>

        {/* Main */}
        <main className="flex-1 p-6 md:p-10">
          <div className="max-w-6xl mx-auto">
            {/* Header */}
            <div className="flex flex-col justify-between gap-4 mb-8 sm:flex-row sm:items-center">
              <div>
                <h2 className="text-3xl font-bold text-gray-900">
                  Tasks
                </h2>
                <p className="mt-1 text-gray-500">
                  Action items extracted from your meetings
                </p>
              </div>
            </div>

            {/* Stats */}
            <div className="grid grid-cols-1 gap-4 mb-8 sm:grid-cols-3">
              <div className="p-5 bg-white border border-gray-200 rounded-xl">
                <p className="text-sm text-gray-500">
                  Total Tasks
                </p>
                <p className="mt-2 text-3xl font-bold text-gray-900">
                  {tasks.length}
                </p>
              </div>

              <div className="p-5 bg-white border border-gray-200 rounded-xl">
                <p className="text-sm text-gray-500">
                  Pending
                </p>
                <p className="mt-2 text-3xl font-bold text-orange-500">
                  {pendingCount}
                </p>
              </div>

              <div className="p-5 bg-white border border-gray-200 rounded-xl">
                <p className="text-sm text-gray-500">
                  Completed
                </p>
                <p className="mt-2 text-3xl font-bold text-green-600">
                  {completedCount}
                </p>
              </div>
            </div>

            {/* Filters */}
            <div className="flex gap-2 mb-6">
              {["all", "pending", "completed"].map((option) => (
                <button
                  key={option}
                  onClick={() => setFilter(option)}
                  className={`px-4 py-2 rounded-lg text-sm font-medium capitalize ${
                    filter === option
                      ? "bg-blue-600 text-white"
                      : "bg-white text-gray-600 border border-gray-200 hover:bg-gray-50"
                  }`}
                >
                  {option}
                </button>
              ))}
            </div>

            {/* Tasks */}
            <div className="overflow-hidden bg-white border border-gray-200 rounded-xl">
              {loading ? (
                <div className="p-10 text-center text-gray-500">
                  Loading tasks...
                </div>
              ) : filteredTasks.length === 0 ? (
                <div className="p-10 text-center">
                  <p className="text-lg font-medium text-gray-700">
                    No tasks found
                  </p>
                  <p className="mt-1 text-sm text-gray-500">
                    Action items from analyzed meetings will appear here.
                  </p>
                </div>
              ) : (
                <div className="divide-y divide-gray-100">
                  {filteredTasks.map((task) => (
                    <div
                      key={task.id}
                      className="flex flex-col gap-4 p-5 transition hover:bg-gray-50 md:flex-row md:items-center md:justify-between"
                    >
                      <div className="flex items-start gap-4">
                        <button
                          onClick={() =>
                            updateTaskStatus(
                              task.id,
                              task.status === "completed"
                                ? "pending"
                                : "completed"
                            )
                          }
                          className={`flex-shrink-0 w-6 h-6 mt-1 border-2 rounded-full flex items-center justify-center ${
                            task.status === "completed"
                              ? "bg-green-500 border-green-500 text-white"
                              : "border-gray-300 hover:border-blue-500"
                          }`}
                        >
                          {task.status === "completed" && "✓"}
                        </button>

                        <div>
                          <p
                            className={`font-medium ${
                              task.status === "completed"
                                ? "text-gray-400 line-through"
                                : "text-gray-900"
                            }`}
                          >
                            {task.task}
                          </p>

                          <div className="flex flex-wrap gap-3 mt-2 text-sm text-gray-500">
                            {task.assignee && (
                              <span>
                                Assignee: {task.assignee}
                              </span>
                            )}

                            {task.deadline && (
                              <span>
                                Deadline: {task.deadline}
                              </span>
                            )}
                          </div>
                        </div>
                      </div>

                      <div className="flex items-center gap-3 md:flex-shrink-0">
                        <span
                          className={`px-3 py-1 text-xs font-medium rounded-full ${
                            task.status === "completed"
                              ? "bg-green-100 text-green-700"
                              : "bg-orange-100 text-orange-700"
                          }`}
                        >
                          {task.status}
                        </span>

                        <Link
                          href={`/meetings/${task.meeting_id}`}
                          className="text-sm font-medium text-blue-600 hover:text-blue-800"
                        >
                          View Meeting
                        </Link>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}