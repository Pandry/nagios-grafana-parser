FROM scratch
# Compile via CGO_ENABLED=0 go build -ldflags '-w -extldflags "-static"'
ADD ./nagios-grafana-backend /nagios-grafana-backend
ENTRYPOINT ["/nagios-grafana-backend"]
