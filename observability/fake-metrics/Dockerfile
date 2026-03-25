FROM python:3.11-slim

RUN pip install --no-cache-dir prometheus-client==0.20.0

WORKDIR /app
COPY app.py .

EXPOSE 8081
CMD ["python3", "app.py"]
