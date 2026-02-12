import os
import json
import logging
from contextlib import asynccontextmanager
from typing import List, Dict

import pika
from fastapi import FastAPI, HTTPException
from transformers import pipeline, AutoTokenizer, AutoModelForSequenceClassification

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Global variables
sentiment_analyzer = None
emotion_classifier = None
keyword_extractor = None
rabbitmq_connection = None
rabbitmq_channel = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Startup and shutdown events"""
    global sentiment_analyzer, emotion_classifier, keyword_extractor
    global rabbitmq_connection, rabbitmq_channel
    
    # Startup
    logger.info("Starting NLP service...")
    
    # Load models
    logger.info("Loading sentiment analysis model...")
    sentiment_analyzer = pipeline(
        "sentiment-analysis",
        model="distilbert-base-uncased-finetuned-sst-2-english",
        device=-1  # Use CPU
    )
    logger.info("Sentiment analysis model loaded")
    
    logger.info("Loading emotion classification model...")
    emotion_classifier = pipeline(
        "text-classification",
        model="j-hartmann/emotion-english-distilroberta-base",
        device=-1,
        top_k=None
    )
    logger.info("Emotion classification model loaded")
    
    # Simple keyword extraction using TF-IDF (for production, use better models)
    logger.info("NLP models loaded successfully")
    
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
    rabbitmq_channel.queue_declare(queue='nlp_queue', durable=True)
    logger.info("RabbitMQ connection established")
    
    # Start consuming messages
    rabbitmq_channel.basic_consume(
        queue='nlp_queue',
        on_message_callback=process_nlp_task
    )
    
    import threading
    def consume():
        logger.info("Starting to consume NLP tasks...")
        rabbitmq_channel.start_consuming()
    
    consumer_thread = threading.Thread(target=consume, daemon=True)
    consumer_thread.start()
    
    yield
    
    # Shutdown
    logger.info("Shutting down NLP service...")
    if rabbitmq_connection and not rabbitmq_connection.is_closed:
        rabbitmq_connection.close()

app = FastAPI(
    title="NLP Service",
    description="Natural Language Processing service for emotion analysis",
    version="1.0.0",
    lifespan=lifespan
)

def extract_keywords(text: str, max_keywords: int = 10) -> List[str]:
    """Extract keywords from text using simple frequency analysis"""
    # Simple keyword extraction (in production, use better models like KeyBERT)
    import re
    from collections import Counter
    
    # Remove special characters and convert to lowercase
    words = re.findall(r'\b[a-zA-Z]{4,}\b', text.lower())
    
    # Common stopwords (simplified list)
    stopwords = {
        'this', 'that', 'with', 'from', 'have', 'been', 'were', 'was',
        'are', 'the', 'and', 'for', 'not', 'but', 'can', 'will', 'would',
        'should', 'could', 'has', 'had', 'does', 'did', 'what', 'when',
        'where', 'who', 'which', 'how', 'why', 'their', 'there', 'they',
        'them', 'then', 'than', 'some', 'such', 'only', 'very', 'just'
    }
    
    # Filter stopwords
    words = [w for w in words if w not in stopwords]
    
    # Get most common words
    word_counts = Counter(words)
    keywords = [word for word, count in word_counts.most_common(max_keywords)]
    
    return keywords

def generate_summary(text: str, max_length: int = 200) -> str:
    """Generate a simple summary by taking first sentences"""
    # Simple summarization (in production, use models like BART or T5)
    sentences = text.split('.')
    summary = ''
    
    for sentence in sentences:
        if len(summary) + len(sentence) < max_length:
            summary += sentence.strip() + '. '
        else:
            break
    
    return summary.strip() or text[:max_length]

def process_nlp_task(ch, method, properties, body):
    """Process an NLP task from the queue"""
    try:
        message = json.loads(body)
        entry_id = message['entry_id']
        transcription = message['transcription']
        
        logger.info(f"Processing NLP analysis for entry: {entry_id}")
        
        # Analyze sentiment
        sentiment_result = sentiment_analyzer(transcription[:512])[0]
        sentiment = sentiment_result['label'].lower()  # positive or negative
        
        # Classify emotions
        emotion_results = emotion_classifier(transcription[:512])[0]
        # Get top 3 emotions
        emotions = sorted(emotion_results, key=lambda x: x['score'], reverse=True)[:3]
        emotion_labels = [e['label'] for e in emotions if e['score'] > 0.1]
        
        # Extract keywords
        keywords = extract_keywords(transcription)
        
        # Generate summary
        summary = generate_summary(transcription)
        
        logger.info(f"NLP analysis completed for entry: {entry_id}")
        logger.info(f"Sentiment: {sentiment}, Emotions: {emotion_labels}")
        
        # Send update to diary service
        update_message = {
            "entry_id": entry_id,
            "sentiment": sentiment,
            "emotions": emotion_labels,
            "keywords": keywords,
            "summary": summary
        }
        
        rabbitmq_channel.queue_declare(queue='nlp_completed', durable=True)
        rabbitmq_channel.basic_publish(
            exchange='',
            routing_key='nlp_completed',
            body=json.dumps(update_message),
            properties=pika.BasicProperties(delivery_mode=2)
        )
        
        logger.info(f"Sent NLP update for entry: {entry_id}")
        
        # Acknowledge message
        ch.basic_ack(delivery_tag=method.delivery_tag)
        
    except Exception as e:
        logger.error(f"Error processing NLP task: {e}", exc_info=True)
        # Reject and requeue message
        ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "sentiment_analyzer": "loaded" if sentiment_analyzer else "not loaded",
        "emotion_classifier": "loaded" if emotion_classifier else "not loaded",
        "rabbitmq": "connected" if rabbitmq_connection and not rabbitmq_connection.is_closed else "disconnected"
    }

@app.post("/analyze")
async def analyze_text(entry_id: str, transcription: str):
    """Manually trigger NLP analysis (for testing)"""
    try:
        message = {
            "entry_id": entry_id,
            "transcription": transcription
        }
        
        rabbitmq_channel.basic_publish(
            exchange='',
            routing_key='nlp_queue',
            body=json.dumps(message),
            properties=pika.BasicProperties(delivery_mode=2)
        )
        
        return {"message": "NLP analysis task queued", "entry_id": entry_id}
    
    except Exception as e:
        logger.error(f"Error queuing NLP analysis: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8001"))
    uvicorn.run(app, host="0.0.0.0", port=port)
