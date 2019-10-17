package plugin

const (
	filebeatConfigTpl = `filebeat.inputs:
- type: docker
  containers.ids: 
  - '*'                                                   // TODO

filebeat.config:
  modules:
    path: ${path.config}/modules.d/*.yml
    reload.enabled: false

fields:
    info:
        launch_type: bpm
        node_xid: {{ .ID }}
        project: development
        protocol_type: POLKADOT
        network_type: public
        user_id: TODO
        environment: {{ .Environment }}
fields_under_root: true
output:
    logstash:
        hosts:
        - "{{ .Collection.Host }}"
        ssl:
            certificate: /etc/ssl/beats/beat.crt
            certificate_authorities:
            - /etc/ssl/beats/ca.crt
            key: /etc/ssl/beats/beat.key
`
)
