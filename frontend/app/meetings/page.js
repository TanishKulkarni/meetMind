"use client";

import { useEffect, useState } from "react";

const API_URL = "http://localhost:8080";

export default function MeetingsPage() {
  const [meetings, setMeetings] = useState([]);
  const [filteredMeetings, setFilteredMeetings] = useState([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [error, setError] = useState("");

  useEffect(() => {
    fetchMeetings();
  }, []);

  useEffect(() => {
    filterMeetings();
  }, [meetings, search, statusFilter]);

  const fetchMeetings = async () => {
    try {
      const response = await fetch(`${API_URL}/api/meetings`);

      if (!response.ok) {
        throw new Error("Failed to fetch meetings");
      }

      const data = await response.json();

      setMeetings(data);
      setError("");
    } catch (error) {
      console.error(error);
      setError("Unable to load meetings.");
    } finally {
      setLoading(false);
    }
  };

  const filterMeetings = () => {
    let result = [...meetings];

    if (search.trim()) {
      const searchTerm = search.toLowerCase();

      result = result.filter(
        (meeting) =>
          meeting.title?.toLowerCase().includes(searchTerm) ||
          meeting.description?.toLowerCase().includes(searchTerm) ||
          meeting.host_name?.toLowerCase().includes(searchTerm)
      );
    }

    if (statusFilter !== "all") {
      result = result.filter(
        (meeting) => meeting.status === statusFilter
      );
    }

    setFilteredMeetings(result);
  };

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
              className="block rounded-lg px-4 py-3 text-sm font-medium text-slate-600 hover:bg-slate-100"
            >
              Dashboard
            </a>

            <a
              href="/meetings"
              className="block rounded-lg bg-blue-50 px-4 py-3 text-sm font-medium text-blue-700"
            >
              Meetings
            </a>

          </nav>

        </aside>

        {/* Main */}
        <section className="flex-1 p-6 md:p-10">

          <div className="flex flex-col justify-between gap-4 md:flex-row md:items-center">

            <div>
              <p className="text-sm font-medium text-blue-600">
                Meetings
              </p>

              <h2 className="mt-1 text-3xl font-bold tracking-tight">
                All Meetings
              </h2>

              <p className="mt-2 text-slate-500">
                View and manage your meeting history.
              </p>
            </div>

            <a
              href="/meetings/new"
              className="rounded-lg bg-blue-600 px-5 py-3 text-sm font-semibold text-white shadow-sm transition hover:bg-blue-700"
            >
              + New Meeting
            </a>

          </div>

          {/* Search + Filter */}
          <div className="mt-8 flex flex-col gap-3 md:flex-row">

            <input
              type="text"
              placeholder="Search meetings..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full rounded-lg border border-slate-200 bg-white px-4 py-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100 md:max-w-md"
            />

            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="rounded-lg border border-slate-200 bg-white px-4 py-3 text-sm outline-none focus:border-blue-500"
            >
              <option value="all">All Statuses</option>
              <option value="completed">Completed</option>
              <option value="uploaded">Uploaded</option>
              <option value="transcribed">Transcribed</option>
              <option value="transcribing">Transcribing</option>
              <option value="analyzing">Analyzing</option>
              <option value="failed">Failed</option>
            </select>

          </div>

          {error && (
            <div className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">
              {error}
            </div>
          )}

          {/* Meetings */}
          <div className="mt-8 overflow-hidden rounded-xl border border-slate-200 bg-white">

            {loading ? (
              <div className="p-10 text-center text-sm text-slate-500">
                Loading meetings...
              </div>
            ) : filteredMeetings.length === 0 ? (
              <div className="p-12 text-center">

                <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-slate-100">
                  📅
                </div>

                <h3 className="mt-4 font-semibold">
                  No meetings found
                </h3>

                <p className="mt-2 text-sm text-slate-500">
                  Try changing your search or status filter.
                </p>

              </div>
            ) : (
              <div>

                {filteredMeetings.map((meeting) => (
                  <a
                    key={meeting.id}
                    href={`/meetings/${meeting.id}`}
                    className="block border-b border-slate-100 p-5 transition last:border-b-0 hover:bg-slate-50"
                  >

                    <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">

                      <div className="min-w-0">

                        <h3 className="truncate font-semibold">
                          {meeting.title}
                        </h3>

                        {meeting.description && (
                          <p className="mt-1 line-clamp-1 text-sm text-slate-500">
                            {meeting.description}
                          </p>
                        )}

                        <div className="mt-3 flex flex-wrap gap-4 text-xs text-slate-400">

                          <span>
                            Host: {meeting.host_name || "Unknown"}
                          </span>

                          <span>
                            {formatDate(meeting.scheduled_at)}
                          </span>

                        </div>

                      </div>

                      <div className="flex items-center gap-4">

                        <StatusBadge status={meeting.status} />

                        <span className="text-slate-400">
                          →
                        </span>

                      </div>

                    </div>

                  </a>
                ))}

              </div>
            )}

          </div>

          {!loading && filteredMeetings.length > 0 && (
            <p className="mt-4 text-sm text-slate-400">
              Showing {filteredMeetings.length} of {meetings.length} meetings
            </p>
          )}

        </section>

      </div>

    </main>
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
      className={`whitespace-nowrap rounded-full px-3 py-1 text-xs font-medium ${style}`}
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