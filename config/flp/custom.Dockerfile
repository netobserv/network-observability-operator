# final stage
FROM quay.io/netobserv/flowlogs-pipeline:latest

COPY config/flp/flowlogs-pipeline.conf.yaml /app/contrib/kubernetes/flowlogs-pipeline.conf.yaml

# expose ports
EXPOSE 2055
EXPOSE 9102

ENTRYPOINT "/app/flowlogs-pipeline" "--config" "/app/contrib/kubernetes/flowlogs-pipeline.conf.yaml"
