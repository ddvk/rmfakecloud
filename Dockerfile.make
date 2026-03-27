FROM debian:bookworm-slim
EXPOSE 3000

# Install Cairo runtime libraries
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libcairo2 \
    && rm -rf /var/lib/apt/lists/*

#ENV RMAPI_HWR_HMAC
#ENV RM_SMTP_SERVER=""
#ENV RM_SMTP_USERNAME=""
#ENV RM_SMTP_PASSWORD=""
COPY dist/rmfakecloud-docker .
ENTRYPOINT ["/rmfakecloud-docker"]
