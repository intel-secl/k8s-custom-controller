#/*
#Copyright © 2019 Intel Corporation
#SPDX-License-Identifier: BSD-3-Clause
#*/

# ISecL K8S Custom Controller
# Applies labels as per Custom Resource Definitions

DESCRIPTION="ISecL K8S Custom Controller"

SERVICE=isecl-k8s-controller
SYSTEMINSTALLDIR=/opt/isecl-k8s-extensions/bin/
SERVICEINSTALLDIR=/etc/systemd/system/
SERVICECONFIG=${SERVICE}.service

VERSION := 1.0-SNAPSHOT
BUILD := `date +%FT%T%z`

# LDFLAGS
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"


# Generate the service binary and executable
.DEFAULT_GOAL: $(SERVICE)
$(SERVICE):
	go build ${LDFLAGS} -o ${SERVICE}-${VERSION} ${SOURCES}

# Install the service binary and the service config files
.PHONY: install
install:
	@mkdir -p ${SYSTEMINSTALLDIR} 
	@cp -f ${SERVICE}-${VERSION} ${SYSTEMINSTALLDIR}
	@cp -f ${SERVICECONFIG} ${SERVICEINSTALLDIR}

# Uninstalls the service binary and the service config files
.PHONY: uninstall
uninstall:
	@service isecl-k8s-controller stop && rm -f ${SERVICEINSTALLDIR}/${SERVICE} ${SERVICEINSTALLDIR}/${SERVICECONFIG}

# Removes the generated service config and binary files
.PHONY: clean
clean:
	@rm -f ${SERVICE}-${VERSION}
