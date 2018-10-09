FROM python:3.7.0-alpine3.8
MAINTAINER NPFLAN

WORKDIR /usr/src/app

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

EXPOSE 9405
COPY exporter.py ./
CMD ["gunicorn", "-b", "0.0.0.0:9405", "-t", "180", "exporter:app"]