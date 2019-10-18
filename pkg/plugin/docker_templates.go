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
-   add_host_metadata: null
-   add_cloud_metadata: null
-   add_docker_metadata: null
-   add_fields:
        fields.log_type: user
        target: ''
        when.or:
        -   equals.container.name: xrp
-   add_fields:
        fields.log_type: system
        target: ''
        when.not.and:
        -   equals.container.name: xrp
-   drop_event.when.not.equals.log_type: user
`
)
