from fastapi import FastAPI, File, UploadFile, HTTPException
from faster_whisper import WhisperModel
import os
import shutil
import tempfile
import requests
import json

OLLAMA_URL = "http://127.0.0.1:11434/api/generate"
QWEN_MODEL = "qwen2.5:3b"


app = FastAPI(
    title="Meet AI Service",
    description="AI service for meeting transcription",
    version="1.0.0",
)


print("Loading Whisper model...")

model = WhisperModel(
    "small",
    device="cpu",
    compute_type="int8",
)

print("Whisper model loaded successfully")


@app.get("/health")
def health_check():
    return {
        "status": "ok",
        "message": "AI service is running",
    }


@app.post("/transcribe")
async def transcribe_audio(file: UploadFile = File(...)):

    if not file.filename:
        raise HTTPException(
            status_code=400,
            detail="Audio file is required",
        )

    temp_path = None

    try:

        file_extension = os.path.splitext(file.filename)[1]

        with tempfile.NamedTemporaryFile(
            delete=False,
            suffix=file_extension,
        ) as temp_file:

            temp_path = temp_file.name

            shutil.copyfileobj(
                file.file,
                temp_file,
            )

        print(f"Transcribing: {file.filename}")

        segments, info = model.transcribe(
            temp_path,
            beam_size=5,
        )

        transcript = " ".join(
            segment.text.strip()
            for segment in segments
        )

        return {
            "filename": file.filename,
            "language": info.language,
            "language_probability": info.language_probability,
            "transcript": transcript,
        }

    except Exception as error:

        print(f"Transcription failed: {error}")

        raise HTTPException(
            status_code=500,
            detail="Transcription failed",
        )

    finally:

        if temp_path and os.path.exists(temp_path):
            os.remove(temp_path)

def analyze_meeting(transcript: str):

    prompt = f"""
You are a meeting intelligence assistant.

Analyze the meeting transcript carefully.

Your task is to extract ONLY information explicitly mentioned in the transcript.

Return ONLY valid JSON.

The JSON MUST follow this exact structure:

{{
    "summary": "short summary of the entire meeting",
    "key_points": [
        "important discussion point 1",
        "important discussion point 2"
    ],
    "decisions": [
        "decision 1",
        "decision 2"
    ],
    "action_items": [
        {{
            "task": "specific task",
            "assignee": "person responsible or null",
            "deadline": "deadline or null"
        }}
    ]
}}

IMPORTANT RULES:

1. Do NOT invent information.
2. Do NOT add information that is not present in the transcript.
3. key_points MUST be an array of strings.
4. decisions MUST be an array of strings.
5. Every concrete task mentioned in the meeting MUST be considered for action_items.
6. If a task has a person responsible for it, put that person's name in assignee.
7. If no person is assigned, use null.
8. If a deadline is mentioned, put it in deadline.
9. If no deadline is mentioned, use null.
10. Keep task descriptions concise.
11. Explicitly mentioned deadlines such as today, tomorrow, Friday, Monday, next week, etc. MUST be captured.
12. Explicitly mentioned people responsible for tasks MUST be captured.
13. Do not confuse discussion points with action items.
14. A decision is something the participants agreed/chose/approved.
15. Return ONLY JSON.
16. Do NOT use markdown.
17. Do NOT include explanations outside the JSON.

MEETING TRANSCRIPT:

{transcript}
"""

    try:

        response = requests.post(
            OLLAMA_URL,
            json={
                "model": QWEN_MODEL,
                "prompt": prompt,
                "stream": False,
                "format": "json",
            },
            timeout=300,
        )

        response.raise_for_status()

        result = response.json()

        return json.loads(result["response"])

    except requests.exceptions.RequestException as error:

        print(f"Ollama request failed: {error}")

        raise HTTPException(
            status_code=500,
            detail="Failed to connect to Qwen",
        )

    except json.JSONDecodeError as error:

        print(f"Invalid JSON from Qwen: {error}")

        raise HTTPException(
            status_code=500,
            detail="Qwen returned invalid JSON",
        )

@app.post("/analyze")
async def analyze_meeting_endpoint(data: dict):

    transcript = data.get("transcript")

    if not transcript:
        raise HTTPException(
            status_code=400,
            detail="Transcript is required",
        )

    print("Analyzing meeting with Qwen...")

    result = analyze_meeting(transcript)

    return {
        "status": "success",
        "analysis": result,
    }