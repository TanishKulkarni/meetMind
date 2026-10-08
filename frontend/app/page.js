"use client";

import { useEffect, useState } from "react";

export default function Home() {
  const [backendStatus, setBackendStatus] = useState("Checking...");
  const [error, setError] = useState("");

  useEffect(() => {
    const checkBackend = async () => {
      try {
        const response = await fetch("http://localhost:8080/api/health");

        if (!response.ok) {
          throw new Error("Backend request failed");
        }

        const data = await response.json();

        setBackendStatus(data.message);
      } catch (error) {
        console.error(error);
        setBackendStatus("Backend unavailable");
        setError("Make sure the Go backend is running on port 8080.");
      }
    };

    checkBackend();
  }, []);

  return (
    <main className="min-h-screen bg-gray-100 flex items-center justify-center p-6">
      <div className="w-full max-w-3xl bg-white rounded-2xl shadow-lg p-8">
        <div className="mb-8">
          <p className="text-sm font-medium text-blue-600">
            AI Meeting Intelligence
          </p>

          <h1 className="text-4xl font-bold text-gray-900 mt-2">
            Meet AI
          </h1>

          <p className="text-gray-600 mt-3">
            AI-powered meeting transcription, summarization and management.
          </p>
        </div>

        <div className="border rounded-xl p-5">
          <h2 className="text-lg font-semibold text-gray-900">
            Backend Status
          </h2>

          <p className="mt-2 text-gray-600">
            {backendStatus}
          </p>

          {error && (
            <p className="mt-3 text-sm text-red-600">
              {error}
            </p>
          )}
        </div>
      </div>
    </main>
  );
}