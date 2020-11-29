FROM scratch
EXPOSE 3000
#ENV RMAPI_HWR_HMAC
#ENV RM_SMTP_SERVER=""
#ENV RM_SMTP_USERNAME=""
#ENV RM_SMTP_PASSWORD=""
COPY dist/rmfakecloud-docker .
ENTRYPOINT ["/rmfakecloud-docker"]
