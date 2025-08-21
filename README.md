# ILLiad PDF upload service
This is a service that bridges virgo and the ILLiad filesystem. It allows Virgo
to upload PDF files for remediation.

### System Requirements
* GO version 1.24 or greater

### Current API

* GET /version : return service version info
* GET /healthcheck : test health of system components; results returned as JSON.
* POST /upload : returns Prometheus metrics

