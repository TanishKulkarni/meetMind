"use client";

import { useEffect, useState } from "react";

const API_URL = "http://localhost:8080";

export default function MeetingDetailsPage({ params }) {
  const [meeting, setMeeting] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    loadMeeting();
  }, []);

  const loadMeeting = async () => {
    try {
      const { id } = await params;

      const response = await fetch(`${API_URL}/api/meetings/${id}`);

      if (!response.ok) {
        if (response.status === 404) {
          throw new Error("Meeting not found.");
        }

        throw new Error("Failed to fetch meeting.");
      }

      const data = await response.json();

      setMeeting(data);
    } catch (error) {
      console.error(error);
      setError(error.message || "Unable to load meeting.");
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-slate-50">
        <p className="text-sm text-slate-500">
          Loading meeting...
        </p>
      </main>
    );
  }

  if (error) {
    return (
      <main className="min-h-screen bg-slate-50 p-6">
        <div className="mx-auto max-w-4xl">
          <a
            href="/meetings"
            className="text-sm font-medium text-blue-600 hover:text-blue-700"
          >
            ← Back to Meetings
          </a>

          <div className="mt-8 rounded-xl border border-red-200 bg-red-50 p-6">
            <h2 className="font-semibold text-red-800">
              Unable to load meeting
            </h2>

            <p className="mt-2 text-sm text-red-600">
              {error}
            </p>
          </div>
        </div>
      </main>
    );
  }

  const analysis = meeting.analysis || {};

  const keyPoints = Array.isArray(analysis.key_points)
    ? analysis.key_points
    : [];

  const decisions = Array.isArray(analysis.decisions)
    ? analysis.decisions
    : [];

  const actionItems = Array.isArray(analysis.action_items)
    ? analysis.action_items
    : [];

  return (
    <main className="min-h-screen bg-slate-50 text-slate-900">
      <div className="mx-auto max-w-6xl p-6 md:p-10">

        {/* Back */}
        <a
          href="/meetings"
          className="text-sm font-medium text-blue-600 hover:text-blue-700"
        >
          ← Back to Meetings
        </a>

        {/* Header */}
        <div className="mt-6 flex flex-col justify-between gap-5 md:flex-row md:items-start">

          <div>
            <div className="flex flex-wrap items-center gap-3">

              <h1 className="text-3xl font-bold tracking-tight">
                {meeting.title}
              </h1>

              <StatusBadge status={meeting.status} />

            </div>

            {meeting.description && (
              <p className="mt-3 max-w-3xl text-slate-500">
                {meeting.description}
              </p>
            )}

            <div className="mt-4 flex flex-wrap gap-5 text-sm text-slate-500">

              <span>
                Host:{" "}
                <span className="font-medium text-slate-700">
                  {meeting.host_name || "Unknown"}
                </span>
              </span>

              <span>
                {formatDate(meeting.scheduled_at)}
              </span>

            </div>
          </div>

        </div>

        {/* Summary */}
        <section className="mt-8 rounded-xl border border-slate-200 bg-white p-6">

          <SectionTitle
            title="Summary"
            description="AI-generated overview of the meeting"
          />

          {analysis.summary ? (
            <p className="mt-5 leading-7 text-slate-600">
              {analysis.summary}
            </p>
          ) : (
            <EmptySection text="No summary available yet." />
          )}

        </section>

        {/* Key Points + Decisions */}
        <div className="mt-6 grid gap-6 lg:grid-cols-2">

          {/* Key Points */}
          <section className="rounded-xl border border-slate-200 bg-white p-6">

            <SectionTitle
              title="Key Points"
              description="Important topics discussed"
            />

            {keyPoints.length > 0 ? (
              <ul className="mt-5 space-y-3">
                {keyPoints.map((point, index) => (
                  <li
                    key={index}
                    className="flex gap-3 text-sm leading-6 text-slate-600"
                  >
                    <span className="mt-2 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-blue-500" />

                    <span>{point}</span>
                  </li>
                ))}
              </ul>
            ) : (
              <EmptySection text="No key points available." />
            )}

          </section>

          {/* Decisions */}
          <section className="rounded-xl border border-slate-200 bg-white p-6">

            <SectionTitle
              title="Decisions"
              description="Agreements and decisions made"
            />

            {decisions.length > 0 ? (
              <ul className="mt-5 space-y-3">
                {decisions.map((decision, index) => (
                  <li
                    key={index}
                    className="flex gap-3 text-sm leading-6 text-slate-600"
                  >
                    <span className="mt-1 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-green-50 text-xs font-semibold text-green-600">
                      ✓
                    </span>

                    <span>{decision}</span>
                  </li>
                ))}
              </ul>
            ) : (
              <EmptySection text="No decisions recorded." />
            )}

          </section>

        </div>

        {/* Action Items */}
        <section className="mt-6 rounded-xl border border-slate-200 bg-white p-6">

          <SectionTitle
            title="Action Items"
            description="Tasks identified from the meeting"
          />

          {actionItems.length > 0 ? (
            <div className="mt-5 overflow-x-auto">

              <table className="w-full min-w-[600px] text-left">

                <thead>
                  <tr className="border-b border-slate-200 text-xs uppercase tracking-wide text-slate-400">

                    <th className="pb-3 pr-6 font-medium">
                      Task
                    </th>

                    <th className="pb-3 pr-6 font-medium">
                      Assignee
                    </th>

                    <th className="pb-3 font-medium">
                      Deadline
                    </th>

                  </tr>
                </thead>

                <tbody>
                  {actionItems.map((item, index) => (
                    <tr
                      key={index}
                      className="border-b border-slate-100 last:border-0"
                    >

                      <td className="py-4 pr-6 text-sm font-medium text-slate-700">
                        {item.task || "—"}
                      </td>

                      <td className="py-4 pr-6 text-sm text-slate-500">
                        {item.assignee || "Unassigned"}
                      </td>

                      <td className="py-4 text-sm text-slate-500">
                        {item.deadline || "No deadline"}
                      </td>

                    </tr>
                  ))}
                </tbody>

              </table>

            </div>
          ) : (
            <EmptySection text="No action items recorded." />
          )}

        </section>

        {/* Transcript */}
        <section className="mt-6 rounded-xl border border-slate-200 bg-white p-6">

          <SectionTitle
            title="Transcript"
            description="Original meeting transcription"
          />

          {meeting.transcript ? (
            <div className="mt-5 max-h-[500px] overflow-y-auto rounded-lg bg-slate-50 p-5">

              <p className="whitespace-pre-wrap text-sm leading-7 text-slate-600">
                {meeting.transcript}
              </p>

            </div>
          ) : (
            <EmptySection text="No transcript available." />
          )}

        </section>

      </div>
    </main>
  );
}

function SectionTitle({ title, description }) {
  return (
    <div>
      <h2 className="text-lg font-semibold text-slate-900">
        {title}
      </h2>

      <p className="mt-1 text-sm text-slate-400">
        {description}
      </p>
    </div>
  );
}

function EmptySection({ text }) {
  return (
    <p className="mt-5 text-sm text-slate-400">
      {text}
    </p>
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