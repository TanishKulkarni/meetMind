from fastapi import FastAPI, File, UploadFile, HTTPException
from faster_whisper import WhisperModel
import os
import shutil
import tempfile


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