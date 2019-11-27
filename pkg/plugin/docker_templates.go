package plugin

const (
	filebeatConfigTpl = `filebeat.inputs:
- type: container
  paths: 
  - '/var/lib/docker/containers/*/*.log'
fields:
    node:
        launch_type: bpm
        xid: {{ .Node.ID }}
fields_under_root: true
output:
{{- if .Node.Collection.Host }}
    logstash:
        hosts:
        - "{{ .Node.Collection.Host }}"
        ssl:
            certificate: /etc/ssl/beats/beat.crt
            certificate_authorities:
            - /etc/ssl/beats/ca.crt
            key: /etc/ssl/beats/beat.key
{{- else }}
    console:
        pretty: true
{{- end }}
`
)
