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
        protocol_type: {{ .Node.Subtype }}
        network_type: {{ .Node.NetworkType }}
        environment: {{ .Node.Environment }}
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
processors:
- drop_event:
        when:
            not:
                or:
                    {{- range .Data.Containers }}
                    {{- if .CollectLogs }}
                    - equals.container.name: {{ $.Node.NamePrefix }}{{ .Name }}
                    {{- end }}
                    {{- end }}

`
)
