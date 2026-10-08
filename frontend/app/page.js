"use client";

import { useEffect, useState } from "react";

const API_URL = "http://localhost:8080";

export default function Home() {
  const [meetings, setMeetings] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    fetchMeetings();
  }, []);

  const fetchMeetings = async () => {
    try {
      setLoading(true);

      const response = await fetch(`${API_URL}/api/meetings`);

      if (!response.ok) {
        throw new Error("Failed to fetch meetings");
      }

      const data = await response.json();

      setMeetings(data);
      setError("");
    } catch (error) {
      console.error(error);
      setError("Unable to connect to the Meet AI backend.");
    } finally {
      setLoading(false);
    }
  };

  const totalMeetings = meetings.length;

  const completedMeetings = meetings.filter(
    (meeting) => meeting.status === "completed"
  ).length;

  const processingMeetings = meetings.filter(
    (meeting) =>
      meeting.status === "processing" ||
      meeting.status === "transcribing" ||
      meeting.status === "analyzing"
  ).length;

  const upcomingMeetings = meetings.filter(
    (meeting) => new Date(meeting.scheduled_at) > new Date()
  ).length;

  return (
    <main className="min-h-screen bg-slate-50 text-slate-900">
      <div className="flex min-h-screen">

        {/* Sidebar */}
        <aside className="hidden w-64 border-r border-slate-200 bg-white p-6 md:block">

          <div className="mb-10">
            <h1 className="text-2xl font-bold tracking-tight">
              Meet<span className="text-blue-600">AI</span>
            </h1>

            <p className="mt-1 text-sm text-slate-500">
              Meeting Intelligence
            </p>
          </div>

          <nav className="space-y-2">

            <a
              href="/"
              className="flex items-center rounded-lg bg-blue-50 px-4 py-3 text-sm font-medium text-blue-700"
            >
              Dashboard
            </a>

            <a
              href="/meetings"
              className="flex items-center rounded-lg px-4 py-3 text-sm font-medium text-slate-600 hover:bg-slate-100"
            >
              Meetings
            </a>

          </nav>

          <div className="absolute bottom-6">
            <p className="text-xs text-slate-400">
              Meet AI v1.0
            </p>
          </div>

        </aside>

        {/* Main Content */}
        <section className="flex-1 p-6 md:p-10">

          {/* Header */}
          <div className="flex flex-col justify-between gap-4 md:flex-row md:items-center">

            <div>
              <p className="text-sm font-medium text-blue-600">
                Meeting Intelligence
              </p>

              <h2 className="mt-1 text-3xl font-bold tracking-tight">
                Good afternoon 👋
              </h2>

              <p className="mt-2 text-slate-500">
                Here&apos;s what&apos;s happening with your meetings.
              </p>
            </div>

            <button className="rounded-lg bg-blue-600 px-5 py-3 text-sm font-semibold text-white shadow-sm transition hover:bg-blue-700">
              + New Meeting
            </button>

          </div>

          {/* Error */}
          {error && (
            <div className="mt-8 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">
              {error}
            </div>
          )}

          {/* Stats */}
          <div className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">

            <StatCard
              title="Total Meetings"
              value={totalMeetings}
            />

            <StatCard
              title="Completed"
              value={completedMeetings}
            />

            <StatCard
              title="Processing"
              value={processingMeetings}
            />

            <StatCard
              title="Upcoming"
              value={upcomingMeetings}
            />

          </div>

          {/* Recent Meetings */}
          <div className="mt-10">

            <div className="mb-4 flex items-center justify-between">

              <div>
                <h3 className="text-xl font-semibold">
                  Recent Meetings
                </h3>

                <p className="mt-1 text-sm text-slate-500">
                  Your latest meeting activity
                </p>
              </div>

              <a
                href="/meetings"
                className="text-sm font-medium text-blue-600 hover:text-blue-700"
              >
                View all
              </a>

            </div>

            <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">

              {loading ? (
                <div className="p-8 text-center text-sm text-slate-500">
                  Loading meetings...
                </div>
              ) : meetings.length === 0 ? (
                <div className="p-10 text-center">

                  <p className="text-lg font-medium text-slate-800">
                    No meetings yet
                  </p>

                  <p className="mt-2 text-sm text-slate-500">
                    Create your first meeting to get started.
                  </p>

                </div>
              ) : (
                <div>

                  {meetings.slice(0, 5).map((meeting) => (
                    <a
                      href={`/meetings/${meeting.id}`}
                      key={meeting.id}
                      className="flex flex-col gap-3 border-b border-slate-100 p-5 transition last:border-b-0 hover:bg-slate-50 md:flex-row md:items-center md:justify-between"
                    >

                      <div>

                        <h4 className="font-medium text-slate-900">
                          {meeting.title}
                        </h4>

                        <p className="mt-1 text-sm text-slate-500">
                          {formatDate(meeting.scheduled_at)}
                        </p>

                      </div>

                      <StatusBadge status={meeting.status} />

                    </a>
                  ))}

                </div>
              )}

            </div>

          </div>

        </section>

      </div>
    </main>
  );
}

function StatCard({ title, value }) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-5">

      <p className="text-sm font-medium text-slate-500">
        {title}
      </p>

      <p className="mt-2 text-3xl font-bold text-slate-900">
        {value}
      </p>

    </div>
  );
}

function StatusBadge({ status }) {
  const statusStyles = {
    completed: "bg-green-50 text-green-700",
    processing: "bg-yellow-50 text-yellow-700",
    transcribing: "bg-yellow-50 text-yellow-700",
    analyzing: "bg-yellow-50 text-yellow-700",
    uploaded: "bg-blue-50 text-blue-700",
    transcribed: "bg-blue-50 text-blue-700",
    failed: "bg-red-50 text-red-700",
  };

  const style =
    statusStyles[status] || "bg-slate-100 text-slate-600";

  return (
    <span
      className={`rounded-full px-3 py-1 text-xs font-medium ${style}`}
    >
      {status}
    </span>
  );
}

function formatDate(date) {
  if (!date) {
    return "No date";
  }

  return new Date(date).toLocaleString("en-IN", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}