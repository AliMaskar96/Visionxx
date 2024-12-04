# Start from the official Go 1.22.4 image
FROM golang:1.22.4

# Set the working directory inside the container
WORKDIR /Chatgpt

# Copy all files from the current directory to /app in the container
COPY . .

# Download Go dependencies
RUN go mod download

# Build the Go application
RUN go build -o Chatgpt

# Expose the default application port (update if necessary)
EXPOSE 8080

# Command to run the application
CMD ["./Chatgpt"]

#docker build -t imag2 .
#docker run -dp 8080:8080 imag2 