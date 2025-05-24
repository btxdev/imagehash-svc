# =============================================
# |                BUILD STAGE                |
# =============================================
FROM golang:1.21 as builder

WORKDIR /app
COPY . .

RUN make build

# =============================================
# |               RUNTIME STAGE               |
# =============================================
FROM gcr.io/distroless/base-debian11

COPY --from=builder /app/bin/server /server
COPY --from=builder /app/migrations /migrations
COPY --from=builder /app/api /api

CMD ["/server"]