FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /owl-esports-backend
EXPOSE 8080

# second stage with python3, mathplotlib and numpy
FROM python:3.12
RUN pip3 install numpy scipy pandas matplotlib

WORKDIR /app

COPY --from=builder /owl-esports-backend /app/owl-esports-backend
COPY --from=builder /app/heatmap.py /app/heatmap.py

ENTRYPOINT ["/app/owl-esports-backend"]