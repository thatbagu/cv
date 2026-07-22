FROM alpine:3.21
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Binary is pre-built by CI using the runner's Go (avoids Docker Hub Go version lag)
COPY cv-tui ./cv-tui
COPY assets/ ./assets/
COPY blog_posts/ ./blog_posts/

ENV FRAMES_DIR=/app/assets/index/animation/cat
ENV ASSETS_DIR=/app/assets
ENV BLOG_DIR=/app/blog_posts

EXPOSE 2222

CMD ["/app/cv-tui"]
