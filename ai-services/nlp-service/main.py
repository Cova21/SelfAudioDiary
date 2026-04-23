import os
import re
import json
import base64
import logging
from collections import Counter
from contextlib import asynccontextmanager
from typing import List

import pika
from fastapi import FastAPI, HTTPException
from transformers import pipeline

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

sentiment_analyzer = None
emotion_classifier = None
rabbitmq_connection = None
rabbitmq_channel = None

EMOTION_CANDIDATES = [
    "радость", "грусть", "спокойствие", "тревога", "гнев",
    "удивление", "вдохновение", "усталость", "благодарность", "ностальгия",
]

EMOTION_EMOJI = {
    "радость": "😊",
    "грусть": "😢",
    "спокойствие": "😌",
    "тревога": "😰",
    "гнев": "😠",
    "удивление": "😲",
    "вдохновение": "✨",
    "усталость": "😴",
    "благодарность": "🙏",
    "ностальгия": "💭",
}

SENTIMENT_MAP = {
    "positive": "позитивное",
    "negative": "негативное",
    "neutral": "нейтральное",
}

RU_STOPWORDS = {
    "быть", "этот", "весь", "свой", "мочь", "который", "такой", "также",
    "только", "ещё", "один", "уже", "себя", "если", "когда", "чтобы",
    "после", "было", "были", "будет", "более", "очень", "просто", "даже",
    "через", "потом", "этого", "того", "может", "можно", "нужно", "надо",
    "тоже", "какой", "всего", "здесь", "есть", "стал", "сейчас", "меня",
    "него", "ничего", "теперь", "каждый", "между", "вообще",
}


@asynccontextmanager
async def lifespan(app: FastAPI):
    global sentiment_analyzer, emotion_classifier
    global rabbitmq_connection, rabbitmq_channel

    logger.info("Starting NLP service...")

    logger.info("Loading multilingual sentiment model...")
    sentiment_analyzer = pipeline(
        "sentiment-analysis",
        model="cardiffnlp/twitter-xlm-roberta-base-sentiment-multilingual",
        device=-1,
        top_k=None,
    )
    logger.info("Sentiment model loaded")

    logger.info("Loading zero-shot emotion classifier...")
    emotion_classifier = pipeline(
        "zero-shot-classification",
        model="MoritzLaurer/mDeBERTa-v3-base-mnli-xnli",
        device=-1,
    )
    logger.info("Emotion classifier loaded")

    logger.info("NLP models loaded successfully")

    rabbitmq_host = os.getenv("RABBITMQ_HOST", "localhost")
    rabbitmq_port = int(os.getenv("RABBITMQ_PORT", "5672"))
    rabbitmq_user = os.getenv("RABBITMQ_USER", "guest")
    rabbitmq_pass = os.getenv("RABBITMQ_PASS", "guest")

    credentials = pika.PlainCredentials(rabbitmq_user, rabbitmq_pass)
    parameters = pika.ConnectionParameters(
        host=rabbitmq_host,
        port=rabbitmq_port,
        credentials=credentials,
    )

    rabbitmq_connection = pika.BlockingConnection(parameters)
    rabbitmq_channel = rabbitmq_connection.channel()
    rabbitmq_channel.queue_declare(queue='nlp_queue', durable=True)
    logger.info("RabbitMQ connection established")

    rabbitmq_channel.basic_consume(
        queue='nlp_queue',
        on_message_callback=process_nlp_task,
    )

    import threading

    def consume():
        logger.info("Starting to consume NLP tasks...")
        rabbitmq_channel.start_consuming()

    consumer_thread = threading.Thread(target=consume, daemon=True)
    consumer_thread.start()

    yield

    logger.info("Shutting down NLP service...")
    if rabbitmq_connection and not rabbitmq_connection.is_closed:
        rabbitmq_connection.close()


app = FastAPI(
    title="NLP Service",
    description="Multilingual NLP service for emotion analysis",
    version="2.0.0",
    lifespan=lifespan,
)


def extract_keywords_ru(text: str, max_keywords: int = 10) -> List[str]:
    words = re.findall(r'\b[а-яёА-ЯЁa-zA-Z]{4,}\b', text.lower())
    words = [w for w in words if w not in RU_STOPWORDS]
    word_counts = Counter(words)
    return [word for word, _ in word_counts.most_common(max_keywords)]


def generate_summary(text: str, max_length: int = 200) -> str:
    sentences = re.split(r'[.!?]+', text)
    summary = ''
    for sentence in sentences:
        s = sentence.strip()
        if not s:
            continue
        if len(summary) + len(s) + 2 < max_length:
            summary += s + '. '
        else:
            break
    return summary.strip() or text[:max_length]


def analyze_sentiment(text: str) -> str:
    results = sentiment_analyzer(text[:512])[0]
    best = max(results, key=lambda x: x['score'])
    label = best['label'].lower()
    return label


def analyze_emotions(text: str) -> List[str]:
    result = emotion_classifier(
        text[:512],
        candidate_labels=EMOTION_CANDIDATES,
        multi_label=True,
    )

    emotions = []
    for label, score in zip(result['labels'], result['scores']):
        if score > 0.25 and len(emotions) < 4:
            emoji = EMOTION_EMOJI.get(label, "")
            emotions.append(f"{emoji} {label}".strip())

    return emotions if emotions else [f"{EMOTION_EMOJI['спокойствие']} спокойствие"]


def process_nlp_task(ch, method, properties, body):
    try:
        if isinstance(body, bytes):
            body = body.decode('utf-8')

        try:
            decoded = base64.b64decode(body).decode('utf-8')
            message = json.loads(decoded)
        except Exception:
            message = json.loads(body)

        if isinstance(message, str):
            message = json.loads(message)

        entry_id = message['entry_id']
        transcription = message['transcription']

        logger.info(f"Processing NLP analysis for entry: {entry_id}")

        sentiment = analyze_sentiment(transcription)
        emotions = analyze_emotions(transcription)
        keywords = extract_keywords_ru(transcription)
        summary = generate_summary(transcription)

        logger.info(f"NLP completed for entry: {entry_id} | sentiment={sentiment}, emotions={emotions}")

        update_message = {
            "entry_id": entry_id,
            "sentiment": sentiment,
            "emotions": emotions,
            "keywords": keywords,
            "summary": summary,
        }

        rabbitmq_channel.queue_declare(queue='nlp_completed', durable=True)
        rabbitmq_channel.basic_publish(
            exchange='',
            routing_key='nlp_completed',
            body=json.dumps(update_message, ensure_ascii=False),
            properties=pika.BasicProperties(delivery_mode=2),
        )

        logger.info(f"Sent NLP update for entry: {entry_id}")
        ch.basic_ack(delivery_tag=method.delivery_tag)

    except Exception as e:
        logger.error(f"Error processing NLP task: {e}", exc_info=True)
        ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)


@app.get("/health")
async def health_check():
    return {
        "status": "healthy",
        "sentiment_analyzer": "loaded" if sentiment_analyzer else "not loaded",
        "emotion_classifier": "loaded" if emotion_classifier else "not loaded",
        "rabbitmq": "connected" if rabbitmq_connection and not rabbitmq_connection.is_closed else "disconnected",
    }


@app.post("/analyze")
async def analyze_text(entry_id: str, transcription: str):
    try:
        message = {"entry_id": entry_id, "transcription": transcription}
        rabbitmq_channel.basic_publish(
            exchange='',
            routing_key='nlp_queue',
            body=json.dumps(message, ensure_ascii=False),
            properties=pika.BasicProperties(delivery_mode=2),
        )
        return {"message": "NLP analysis task queued", "entry_id": entry_id}
    except Exception as e:
        logger.error(f"Error queuing NLP analysis: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8001"))
    uvicorn.run(app, host="0.0.0.0", port=port)
