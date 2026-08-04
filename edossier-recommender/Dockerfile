FROM python:3.11-slim

LABEL org.opencontainers.image.title="e-dossier-recommender"
LABEL org.opencontainers.image.description="ML-powered school facility shortage detector"
LABEL org.opencontainers.image.version="1.0.0"

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

RUN apt-get update && \
    apt-get install -y --no-install-recommends build-essential && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt && \
    apt-get purge -y --auto-remove build-essential

COPY recommender.py .

RUN useradd -m appuser && chown -R appuser:appuser /app
USER appuser

EXPOSE 9001

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:9001/health')" || exit 1

CMD ["uvicorn", "recommender:app", "--host", "0.0.0.0", "--port", "9001"]
