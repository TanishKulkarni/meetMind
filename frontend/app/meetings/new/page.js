"use client";

import { useState } from "react";

const API_URL = "http://localhost:8080";

export default function NewMeetingPage() {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [hostName, setHostName] = useState("");
  const [hostEmail, setHostEmail] = useState("");
  const [scheduledAt, setScheduledAt] = useState("");
  const [audioFile, setAudioFile] = useState(null);

  const [step, setStep] = useState("form");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const createAndProcessMeeting = async (event) => {
    event.preventDefault();

    if (!title.trim()) {
      setError("Meeting title is required.");
      return;
    }

    if (!scheduledAt) {
      setError("Please select a meeting date and time.");
      return;
    }

    if (!audioFile) {
      setError("Please select an audio file.");
      return;
    }

    try {
      setError("");
      setLoading(true);

      // 1. Create meeting
      setStep("creating");

      const createResponse = await fetch(
        `${API_URL}/api/meetings`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            title,
            description,
            host_name: hostName,
            host_email: hostEmail,
            scheduled_at: new Date(
              scheduledAt
            ).toISOString(),
          }),
        }
      );

      if (!createResponse.ok) {
        const data = await createResponse.json();

        throw new Error(
          data.error || "Failed to create meeting."
        );
      }

      const meeting = await createResponse.json();

      // 2. Upload audio
      setStep("uploading");

      const formData = new FormData();

      formData.append("audio", audioFile);

      const uploadResponse = await fetch(
        `${API_URL}/api/meetings/${meeting.id}/upload`,
        {
          method: "POST",
          body: formData,
        }
      );

      if (!uploadResponse.ok) {
        const data = await uploadResponse.json();

        throw new Error(
          data.error || "Failed to upload audio."
        );
      }

      // 3. Process meeting
      setStep("processing");

      const processResponse = await fetch(
        `${API_URL}/api/meetings/${meeting.id}/process`,
        {
          method: "POST",
        }
      );

      if (!processResponse.ok) {
        const data = await processResponse.json();

        throw new Error(
          data.error || "Failed to process meeting."
        );
      }

      // 4. Open completed meeting
      setStep("completed");

      window.location.href =
        `/meetings/${meeting.id}`;

    } catch (error) {
      console.error(error);

      setError(
        error.message || "Something went wrong."
      );

      setStep("form");
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="min-h-screen bg-slate-50 text-slate-900">

      <div className="mx-auto max-w-3xl p-6 md:p-10">

        {/* Back */}
        <a
          href="/meetings"
          className="text-sm font-medium text-blue-600 hover:text-blue-700"
        >
          ← Back to Meetings
        </a>

        {/* Header */}
        <div className="mt-6">

          <p className="text-sm font-medium text-blue-600">
            New Meeting
          </p>

          <h1 className="mt-1 text-3xl font-bold tracking-tight">
            Create a Meeting
          </h1>

          <p className="mt-2 text-slate-500">
            Upload your meeting recording and let Meet AI
            transcribe and analyze it.
          </p>

        </div>

        <form
          onSubmit={createAndProcessMeeting}
          className="mt-8 space-y-6"
        >

          {/* Meeting Information */}
          <section className="rounded-xl border border-slate-200 bg-white p-6">

            <h2 className="text-lg font-semibold">
              Meeting Information
            </h2>

            <p className="mt-1 text-sm text-slate-400">
              Basic information about the meeting.
            </p>

            <div className="mt-6 space-y-5">

              {/* Title */}
              <div>

                <label className="mb-2 block text-sm font-medium">
                  Meeting Title *
                </label>

                <input
                  type="text"
                  value={title}
                  onChange={(e) =>
                    setTitle(e.target.value)
                  }
                  placeholder="Weekly Team Meeting"
                  disabled={loading}
                  className="w-full rounded-lg border border-slate-200 px-4 py-3 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100 disabled:bg-slate-50"
                />

              </div>

              {/* Description */}
              <div>

                <label className="mb-2 block text-sm font-medium">
                  Description
                </label>

                <textarea
                  value={description}
                  onChange={(e) =>
                    setDescription(e.target.value)
                  }
                  placeholder="Discuss project progress and next steps..."
                  rows={3}
                  disabled={loading}
                  className="w-full resize-none rounded-lg border border-slate-200 px-4 py-3 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100 disabled:bg-slate-50"
                />

              </div>

              {/* Host */}
              <div className="grid gap-5 md:grid-cols-2">

                <div>

                  <label className="mb-2 block text-sm font-medium">
                    Host Name
                  </label>

                  <input
                    type="text"
                    value={hostName}
                    onChange={(e) =>
                      setHostName(e.target.value)
                    }
                    placeholder="Tanish Kulkarni"
                    disabled={loading}
                    className="w-full rounded-lg border border-slate-200 px-4 py-3 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100 disabled:bg-slate-50"
                  />

                </div>

                <div>

                  <label className="mb-2 block text-sm font-medium">
                    Host Email
                  </label>

                  <input
                    type="email"
                    value={hostEmail}
                    onChange={(e) =>
                      setHostEmail(e.target.value)
                    }
                    placeholder="you@example.com"
                    disabled={loading}
                    className="w-full rounded-lg border border-slate-200 px-4 py-3 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100 disabled:bg-slate-50"
                  />

                </div>

              </div>

              {/* Date */}
              <div>

                <label className="mb-2 block text-sm font-medium">
                  Meeting Date & Time *
                </label>

                <input
                  type="datetime-local"
                  value={scheduledAt}
                  onChange={(e) =>
                    setScheduledAt(e.target.value)
                  }
                  disabled={loading}
                  className="w-full rounded-lg border border-slate-200 px-4 py-3 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100 disabled:bg-slate-50"
                />

              </div>

            </div>

          </section>

          {/* Audio */}
          <section className="rounded-xl border border-slate-200 bg-white p-6">

            <h2 className="text-lg font-semibold">
              Meeting Recording
            </h2>

            <p className="mt-1 text-sm text-slate-400">
              Upload the audio recording that you want Meet AI
              to transcribe and analyze.
            </p>

            <div className="mt-6">

              <label
                htmlFor="audio"
                className="flex cursor-pointer flex-col items-center justify-center rounded-xl border-2 border-dashed border-slate-200 p-10 text-center transition hover:border-blue-400 hover:bg-blue-50/30"
              >

                <div className="text-3xl">
                  🎙️
                </div>

                <p className="mt-3 text-sm font-medium">
                  {audioFile
                    ? audioFile.name
                    : "Choose an audio file"}
                </p>

                <p className="mt-1 text-xs text-slate-400">
                  MP3, WAV, M4A and other supported formats
                </p>

                <input
                  id="audio"
                  type="file"
                  accept="audio/*"
                  onChange={(e) =>
                    setAudioFile(
                      e.target.files?.[0] || null
                    )
                  }
                  disabled={loading}
                  className="hidden"
                />

              </label>

            </div>

          </section>

          {/* Error */}
          {error && (
            <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">
              {error}
            </div>
          )}

          {/* Progress */}
          {loading && (
            <div className="rounded-xl border border-blue-200 bg-blue-50 p-5">

              <p className="font-medium text-blue-800">
                {getStepMessage(step)}
              </p>

              <div className="mt-4 h-2 overflow-hidden rounded-full bg-blue-100">

                <div
                  className={`h-full rounded-full bg-blue-600 transition-all ${
                    step === "creating"
                      ? "w-1/4"
                      : step === "uploading"
                      ? "w-2/4"
                      : step === "processing"
                      ? "w-3/4"
                      : "w-full"
                  }`}
                />

              </div>

              {step === "processing" && (
                <p className="mt-3 text-xs text-blue-600">
                  This may take a little while because the
                  audio is being transcribed and analyzed
                  locally.
                </p>
              )}

            </div>
          )}

          {/* Submit */}
          <div className="flex justify-end">

            <button
              type="submit"
              disabled={loading}
              className="rounded-lg bg-blue-600 px-6 py-3 text-sm font-semibold text-white shadow-sm transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {loading
                ? "Processing..."
                : "Create & Analyze Meeting"}
            </button>

          </div>

        </form>

      </div>

    </main>
  );
}

function getStepMessage(step) {
  switch (step) {
    case "creating":
      return "Creating meeting...";

    case "uploading":
      return "Uploading meeting recording...";

    case "processing":
      return "Transcribing and analyzing meeting...";

    case "completed":
      return "Meeting processed successfully.";

    default:
      return "Preparing...";
  }
}