# neuPrintHTTP


[![GitHub issues](https://img.shields.io/github/issues/connectome-neuprint/neuPrintHTTP.svg)](https://GitHub.com/connectome-neuprint/neuPrintHTTP/issues/)

Implements a connectomics REST interface that leverages the [neuprint](https://github.com/janelia-flyem/neuPrint) data model.

## Installation

    % go get github.com/connectome-neuprint/neuPrintHTTP

## Running

    % neuPrintHTTP -port |PORTNUM| config.json
 
The config file should contain information on the backend datastore that satisfies the connectomics REST API and the location for a file containing
a list of authorized users.  To test https locally and generate the necessary certificates, run:

    % go run $GOROOT/src/crypto/tls/generate_cert.go --host localhost

## Installing without kafka support

If you are having trouble building the server, because librdkafka is missing and you don't need to send log messages to a kafka server, then try this build.

    %  go get -tags nokafka github.com/connectome-neuprint/neuPrintHTTP
