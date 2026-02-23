import os
import json
import logging
import tempfile
from contextlib import asynccontextmanager

from faster_whisper import WhisperModel
import pika
from fastapi import FastAPI, HTTPException
from minio import Minio
from minio.error import S3Error

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Global variables
whisper_model = None
minio_client = None
rabbitmq_connection = None
rabbitmq_channel = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Startup and shutdown events"""
    global whisper_model, minio_client, rabbitmq_connection, rabbitmq_channel
    
    # Startup
    logger.info("Starting transcription service...")
    
    # Load Whisper model (faster-whisper uses int8 on CPU for efficiency)
    model_name = os.getenv("WHISPER_MODEL", "base")
    logger.info(f"Loading Whisper model: {model_name}")
    whisper_model = WhisperModel(model_name, device="cpu", compute_type="int8")
    logger.info("Whisper model loaded successfully")
    
    # Initialize MinIO client
    minio_endpoint = os.getenv("MINIO_ENDPOINT", "localhost:9000")
    minio_access_key = os.getenv("MINIO_ACCESS_KEY", "minioadmin")
    minio_secret_key = os.getenv("MINIO_SECRET_KEY", "minioadmin")
    minio_use_ssl = os.getenv("MINIO_USE_SSL", "false").lower() == "true"
    
    minio_client = Minio(
        minio_endpoint,
        access_key=minio_access_key,
        secret_key=minio_secret_key,
        secure=minio_use_ssl
    )
    logger.info("MinIO client initialized")
    
    # Connect to RabbitMQ
    rabbitmq_host = os.getenv("RABBITMQ_HOST", "localhost")
    rabbitmq_port = int(os.getenv("RABBITMQ_PORT", "5672"))
    rabbitmq_user = os.getenv("RABBITMQ_USER", "guest")
    rabbitmq_pass = os.getenv("RABBITMQ_PASS", "guest")
    
    credentials = pika.PlainCredentials(rabbitmq_user, rabbitmq_pass)
    parameters = pika.ConnectionParameters(
        host=rabbitmq_host,
        port=rabbitmq_port,
        credentials=credentials
    )
    
    rabbitmq_connection = pika.BlockingConnection(parameters)
    rabbitmq_channel = rabbitmq_connection.channel()
    rabbitmq_channel.queue_declare(queue='transcription_queue', durable=True)
    logger.info("RabbitMQ connection established")
    
    # Start consuming messages
    rabbitmq_channel.basic_consume(
        queue='transcription_queue',
        on_message_callback=process_transcription_task,
        auto_ack=True
    )
    
    import threading
    def consume():
        logger.info("Starting to consume transcription tasks...")
        rabbitmq_channel.start_consuming()
    
    consumer_thread = threading.Thread(target=consume, daemon=True)
    consumer_thread.start()
    
    yield
    
    # Shutdown
    logger.info("Shutting down transcription service...")
    if rabbitmq_connection and not rabbitmq_connection.is_closed:
        rabbitmq_connection.close()

app = FastAPI(
    title="Transcription Service",
    description="Audio transcription service using OpenAI Whisper",
    version="1.0.0",
    lifespan=lifespan
)

def process_transcription_task(ch, method, properties, body):
    """Process a transcription task from the queue"""
    try:
        # Decode if bytes
        if isinstance(body, bytes):
            body = body.decode('utf-8')
        
        logger.info(f"Received message body type: {type(body)}, content: {body[:200] if len(str(body)) > 200 else body}")
        
        # Try to decode from base64 first (RabbitMQ might encode it)
        import base64
        try:
            decoded = base64.b64decode(body).decode('utf-8')
            logger.info(f"Decoded from base64: {decoded}")
            message = json.loads(decoded)
        except Exception as e:
            # If base64 decode fails, try direct JSON parse
            logger.info(f"Base64 decode failed, trying direct JSON parse: {e}")
            message = json.loads(body)
        
        logger.info(f"Parsed message type: {type(message)}, content: {message}")
        
        entry_id = message['entry_id']
        user_id = message['user_id']
        audio_file_id = message['audio_file_id']
        
        logger.info(f"Processing transcription for entry: {entry_id}")
        
        # Download audio file from MinIO
        bucket_name = os.getenv("MINIO_BUCKET", "voice-diary-audio")
        object_name = f"{user_id}/{audio_file_id}"
        
        # Create temporary file
        with tempfile.NamedTemporaryFile(suffix=".audio", delete=False) as temp_file:
            temp_path = temp_file.name
            
            try:
                minio_client.fget_object(bucket_name, object_name, temp_path)
                logger.info(f"Downloaded audio file: {object_name}")
                
                # Transcribe audio (faster-whisper API)
                seg_iter, info = whisper_model.transcribe(
                    temp_path,
                    language="ru",
                    beam_size=5,
                    initial_prompt="Личный голосовой дневник на разговорном русском языке."
                )

                segments = []
                transcription_parts = []
                for segment in seg_iter:
                    transcription_parts.append(segment.text)
                    segments.append({
                        "start_time_ms": int(segment.start * 1000),
                        "end_time_ms": int(segment.end * 1000),
                        "text": segment.text
                    })

                transcription = " ".join(transcription_parts)
                language = info.language
                
                logger.info(f"Transcription completed for entry: {entry_id}")
                
                # Update diary entry via gRPC (simplified - in production use proper gRPC client)
                # For now, we'll send a notification via RabbitMQ
                update_message = {
                    "entry_id": entry_id,
                    "user_id": user_id,
                    "transcription": transcription,
                    "segments": segments,
                    "language": language
                }
                
                # Send to diary service update queue
                rabbitmq_channel.queue_declare(queue='transcription_completed', durable=True)
                rabbitmq_channel.basic_publish(
                    exchange='',
                    routing_key='transcription_completed',
                    body=json.dumps(update_message),
                    properties=pika.BasicProperties(delivery_mode=2)
                )
                
                logger.info(f"Sent transcription update for entry: {entry_id}")
                
            finally:
                # Clean up temporary file
                if os.path.exists(temp_path):
                    os.unlink(temp_path)
        
    except Exception as e:
        logger.error(f"Error processing transcription task: {e}", exc_info=True)

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "model": os.getenv("WHISPER_MODEL", "base"),
        "minio": "connected" if minio_client else "disconnected",
        "rabbitmq": "connected" if rabbitmq_connection and not rabbitmq_connection.is_closed else "disconnected"
    }

@app.post("/transcribe")
async def transcribe_audio(entry_id: str, user_id: str, audio_file_id: str):
    """Manually trigger transcription (for testing)"""
    try:
        message = {
            "entry_id": entry_id,
            "user_id": user_id,
            "audio_file_id": audio_file_id
        }
        
        rabbitmq_channel.basic_publish(
            exchange='',
            routing_key='transcription_queue',
            body=json.dumps(message),
            properties=pika.BasicProperties(delivery_mode=2)
        )
        
        return {"message": "Transcription task queued", "entry_id": entry_id}
    
    except Exception as e:
        logger.error(f"Error queuing transcription: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run(app, host="0.0.0.0", port=port)
