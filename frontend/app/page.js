"use client";

import { useEffect, useState } from "react";
import Link from "next/link";

const API_URL = "http://localhost:8080";

export default function Dashboard() {
  const [meetings, setMeetings] = useState([]);
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(true);

  const fetchDashboardData = async () => {
    try {
      const [meetingsResponse, tasksResponse] = await Promise.all([
        fetch(`${API_URL}/api/meetings`),
        fetch(`${API_URL}/api/tasks`),
      ]);

      if (!meetingsResponse.ok || !tasksResponse.ok) {
        throw new Error("Failed to fetch dashboard data");
      }

      const meetingsData = await meetingsResponse.json();
      const tasksData = await tasksResponse.json();

      setMeetings(meetingsData.meetings || []);
      setTasks(tasksData.tasks || []);
    } catch (error) {
      console.error("Dashboard fetch error:", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const completedMeetings = meetings.filter(
    (meeting) => meeting.status === "completed"
  ).length;

  const pendingTasks = tasks.filter(
    (task) => task.status === "pending"
  );

  const recentMeetings = meetings.slice(0, 5);

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
      console.error("Task update error:", error);
    }
  };

  const formatDate = (date) => {
    if (!date) return "No date";

    return new Date(date).toLocaleDateString("en-IN", {
      day: "numeric",
      month: "short",
      year: "numeric",
    });
  };

  const getStatusStyle = (status) => {
    switch (status) {
      case "completed":
        return "bg-green-100 text-green-700";
      case "processing":
      case "analyzing":
      case "transcribing":
        return "bg-blue-100 text-blue-700";
      case "failed":
        return "bg-red-100 text-red-700";
      default:
        return "bg-gray-100 text-gray-700";
    }
  };

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
              className="flex items-center px-4 py-3 font-medium text-blue-600 bg-blue-50 rounded-lg"
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
              className="flex items-center px-4 py-3 text-gray-600 rounded-lg hover:bg-gray-100"
            >
              Tasks
            </Link>
          </nav>
        </aside>

        {/* Main Content */}
        <main className="flex-1 p-6 md:p-10">
          <div className="max-w-7xl mx-auto">

            {/* Header */}
            <div className="flex flex-col justify-between gap-4 mb-8 sm:flex-row sm:items-center">
              <div>
                <h2 className="text-3xl font-bold text-gray-900">
                  Dashboard
                </h2>

                <p className="mt-1 text-gray-500">
                  Overview of your meetings and action items
                </p>
              </div>

              <Link
                href="/meetings/new"
                className="px-5 py-3 font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700"
              >
                + New Meeting
              </Link>
            </div>

            {/* Stats */}
            <div className="grid grid-cols-1 gap-5 mb-8 md:grid-cols-3">

              <div className="p-6 bg-white border border-gray-200 rounded-xl">
                <p className="text-sm font-medium text-gray-500">
                  Total Meetings
                </p>

                <p className="mt-2 text-3xl font-bold text-gray-900">
                  {loading ? "—" : meetings.length}
                </p>
              </div>

              <div className="p-6 bg-white border border-gray-200 rounded-xl">
                <p className="text-sm font-medium text-gray-500">
                  Completed Meetings
                </p>

                <p className="mt-2 text-3xl font-bold text-green-600">
                  {loading ? "—" : completedMeetings}
                </p>
              </div>

              <div className="p-6 bg-white border border-gray-200 rounded-xl">
                <p className="text-sm font-medium text-gray-500">
                  Pending Tasks
                </p>

                <p className="mt-2 text-3xl font-bold text-orange-500">
                  {loading ? "—" : pendingTasks.length}
                </p>
              </div>

            </div>

            {/* Two-column section */}
            <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">

              {/* Recent Meetings */}
              <section className="bg-white border border-gray-200 rounded-xl">

                <div className="flex items-center justify-between p-6 border-b border-gray-100">
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900">
                      Recent Meetings
                    </h3>

                    <p className="mt-1 text-sm text-gray-500">
                      Your latest meetings
                    </p>
                  </div>

                  <Link
                    href="/meetings"
                    className="text-sm font-medium text-blue-600 hover:text-blue-800"
                  >
                    View all
                  </Link>
                </div>

                <div className="divide-y divide-gray-100">

                  {loading ? (
                    <div className="p-6 text-sm text-gray-500">
                      Loading meetings...
                    </div>
                  ) : recentMeetings.length === 0 ? (
                    <div className="p-6 text-sm text-gray-500">
                      No meetings yet.
                    </div>
                  ) : (
                    recentMeetings.map((meeting) => (
                      <Link
                        key={meeting.id}
                        href={`/meetings/${meeting.id}`}
                        className="flex items-center justify-between p-5 hover:bg-gray-50"
                      >
                        <div className="min-w-0">
                          <p className="font-medium text-gray-900 truncate">
                            {meeting.title}
                          </p>

                          <p className="mt-1 text-sm text-gray-500">
                            {formatDate(meeting.scheduled_at)}
                          </p>
                        </div>

                        <span
                          className={`ml-4 px-3 py-1 text-xs font-medium rounded-full capitalize ${getStatusStyle(
                            meeting.status
                          )}`}
                        >
                          {meeting.status}
                        </span>
                      </Link>
                    ))
                  )}

                </div>
              </section>

              {/* Pending Tasks */}
              <section className="bg-white border border-gray-200 rounded-xl">

                <div className="flex items-center justify-between p-6 border-b border-gray-100">
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900">
                      Pending Tasks
                    </h3>

                    <p className="mt-1 text-sm text-gray-500">
                      Action items requiring attention
                    </p>
                  </div>

                  <Link
                    href="/tasks"
                    className="text-sm font-medium text-blue-600 hover:text-blue-800"
                  >
                    View all
                  </Link>
                </div>

                <div className="divide-y divide-gray-100">

                  {loading ? (
                    <div className="p-6 text-sm text-gray-500">
                      Loading tasks...
                    </div>
                  ) : pendingTasks.length === 0 ? (
                    <div className="p-6">
                      <p className="font-medium text-gray-700">
                        No pending tasks 🎉
                      </p>

                      <p className="mt-1 text-sm text-gray-500">
                        You're all caught up.
                      </p>
                    </div>
                  ) : (
                    pendingTasks.slice(0, 5).map((task) => (
                      <div
                        key={task.id}
                        className="flex items-start gap-4 p-5"
                      >
                        <button
                          onClick={() =>
                            updateTaskStatus(task.id, "completed")
                          }
                          className="flex-shrink-0 w-6 h-6 mt-1 border-2 border-gray-300 rounded-full hover:border-green-500"
                          title="Mark as completed"
                        />

                        <div className="min-w-0">
                          <p className="font-medium text-gray-900">
                            {task.task}
                          </p>

                          <div className="flex flex-wrap gap-3 mt-2 text-sm text-gray-500">

                            {task.assignee && (
                              <span>
                                {task.assignee}
                              </span>
                            )}

                            {task.deadline && (
                              <span>
                                Deadline: {task.deadline}
                              </span>
                            )}

                          </div>

                          <Link
                            href={`/meetings/${task.meeting_id}`}
                            className="inline-block mt-2 text-sm font-medium text-blue-600 hover:text-blue-800"
                          >
                            View meeting →
                          </Link>
                        </div>
                      </div>
                    ))
                  )}

                </div>
              </section>

            </div>

            {/* Quick Actions */}
            <section className="mt-6 p-6 bg-white border border-gray-200 rounded-xl">

              <h3 className="text-lg font-semibold text-gray-900">
                Quick Actions
              </h3>

              <div className="grid grid-cols-1 gap-4 mt-4 sm:grid-cols-3">

                <Link
                  href="/meetings/new"
                  className="p-4 border border-gray-200 rounded-lg hover:border-blue-400 hover:bg-blue-50"
                >
                  <p className="font-medium text-gray-900">
                    New Meeting
                  </p>

                  <p className="mt-1 text-sm text-gray-500">
                    Upload and analyze a meeting
                  </p>
                </Link>

                <Link
                  href="/meetings"
                  className="p-4 border border-gray-200 rounded-lg hover:border-blue-400 hover:bg-blue-50"
                >
                  <p className="font-medium text-gray-900">
                    Browse Meetings
                  </p>

                  <p className="mt-1 text-sm text-gray-500">
                    View your meeting history
                  </p>
                </Link>

                <Link
                  href="/tasks"
                  className="p-4 border border-gray-200 rounded-lg hover:border-blue-400 hover:bg-blue-50"
                >
                  <p className="font-medium text-gray-900">
                    Manage Tasks
                  </p>

                  <p className="mt-1 text-sm text-gray-500">
                    Review and complete action items
                  </p>
                </Link>

              </div>
            </section>

          </div>
        </main>
      </div>
    </div>
  );
}