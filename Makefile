# CIT K8S Custom Controller
# Applies labels as per Custom Resource Definitions
# Author:  <manux.ullas@intel.com>
DESCRIPTION="CIT K8S Custom Controller"

SERVICE=citk8scontroller
SYSTEMINSTALLDIR=/opt/cit_k8s_extensions/bin/
SERVICEINSTALLDIR=/etc/systemd/system/
SERVICECONFIG=${SERVICE}.service

VERSION := 1.0-SNAPSHOT
BUILD := `date +%FT%T%z`

# LDFLAGS
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"


# Generate the service binary and executable
.DEFAULT_GOAL: $(SERVICE)
$(SERVICE):
	glide update -v
	go build ${LDFLAGS} -o ${SERVICE}-${VERSION} ${SOURCES}

# Install the service binary and the service config files
.PHONY: install
install:
	@service citk8scontroller stop
	@cp -f ${SERVICE}-${VERSION} ${SYSTEMINSTALLDIR}
	@cp -f ${SERVICECONFIG} ${SERVICEINSTALLDIR}
	@systemctl daemon-reload
	@service citk8scontroller start

# Uninstalls the service binary and the service config files
.PHONY: uninstall
uninstall:
	@service citk8scontroller stop && rm -f ${SERVICEINSTALLDIR}/${SERVICE} ${SERVICEINSTALLDIR}/${SERVICECONFIG}

# Removes the generated service config and binary files
.PHONY: clean
clean:
	@rm -rf vendor/
	@rm -f ${SERVICE}-${VERSION}
