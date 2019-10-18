package plugin

const (
	filebeatConfigTpl = `filebeat.inputs:
- type: container
// filebeat.config:
//   modules:
//     path: ${path.config}/modules.d/*.yml
//     reload.enabled: false
fields:
    info:
        launch_type: bpm
        node_xid: {{ .Node.ID }}
        project: development
        protocol_type: POLKADOT
        network_type: public
        user_id: TODO
        environment: {{ .Node.Environment }}
fields_under_root: true
output:
    logstash:
        hosts:
        - "{{ .Node.Collection.Host }}"
        ssl:
            certificate: /etc/ssl/beats/beat.crt
            certificate_authorities:
            - /etc/ssl/beats/ca.crt
            key: /etc/ssl/beats/beat.key
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
