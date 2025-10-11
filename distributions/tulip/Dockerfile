FROM alpine:latest

LABEL org.opencontainers.image.title="OllyGarden Tulip" \
      org.opencontainers.image.description="OllyGarden Tulip - A commercially supported distribution of the OpenTelemetry Collector with enterprise-grade support and SLAs. Collects, processes, and exports telemetry data (metrics, traces, logs) with a curated set of components optimized for production use." \
      org.opencontainers.image.vendor="OllyGarden" \
      org.opencontainers.image.url="https://github.com/ollygarden/tulip" \
      org.opencontainers.image.documentation="https://github.com/ollygarden/tulip"

ARG USER_UID=10001
USER ${USER_UID}

COPY --chmod=755 tulip /tulip
COPY --chmod=755 entrypoint.sh /entrypoint.sh
COPY config.yaml /etc/tulip/config.yaml
ENTRYPOINT ["/entrypoint.sh"]
CMD []
EXPOSE 4317
