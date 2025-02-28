# Use the official Python base image
FROM python:3.9-slim

# Set the working directory inside the container
WORKDIR /app

# Copy the Python script and any required files into the container
COPY . /app/

# Install required Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Set the entry point to run your Python script
CMD ["python", "consumer.py"]
