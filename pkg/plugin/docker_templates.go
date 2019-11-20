package plugin

const (
	filebeatConfigTpl = `filebeat.inputs:
- type: container
  paths: 
  - '/var/lib/docker/containers/*/*.log'
fields:
    info:
        launch_type: bpm
        node_xid: {{ .Node.ID }}
        protocol_type: {{ .Node.Protocol }}
        network_type: {{ .Node.NetworkType }}
        environment: {{ .Node.Network }}
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
