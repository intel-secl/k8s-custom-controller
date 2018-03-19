# CIT K8S Docker Container Integrity Plugin
# Builds, Installs and Uninstalls the image authorization plugin service
# Author: Manu Ullas <manux.ullas@intel.com>
DESCRIPTION="CIT K8S Docker Container Integrity Plugin"

SERVICE=citk8s-custom-controller
SYSTEMINSTALLDIR=/opt/citk8s/$(SERVICE)

VERSION := 1.0.0
BUILD := `date +%FT%T%z`

# LDFLAGS
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"


# Generate the service binary and executable
.DEFAULT_GOAL: $(SERVICE)
$(SERVICE):
	glide update -v
	go build ${LDFLAGS} -o ${SERVICE} ${SOURCES}

# Install the service binary and the service config files
.PHONY: install
install:
	@cp -f ${SERVICE} ${SERVICEINSTALLDIR}

# Uninstalls the service binary and the service config files
.PHONY: uninstall
uninstall:
	@rm -f ${SERVICEINSTALLDIR}/${SERVICE}

# Removes the generated service config and binary files
.PHONY: clean
clean:
	@rm -rf vendor/
	@rm -f ${SERVICE}
