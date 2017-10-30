##############################################################################
Pre-requisites for building the code
##############################################################################
1. Install GO
2. Install MAVEN
3. Install maven plugin for GO - from https://github.com/raydac/mvn-golang
4. Set GOROOT to path of go folder. For example : /usr/local/go
5. Install JAVA SDK and set JAVA_HOME

Run below command from terminal inside the main directory
-- mvn package
On build success, you can find binary in "bin" folder
Binary name - citk8scontroller-1.0-SNAPSHOT

##############################################################################
Installation of binary
##############################################################################
Pre-requisites
1. Kubernetes cluster should be up and running
2. Copy the binary from bin folder to /opt folder 
	mv  citk8scontroller-1.0-SNAPSHOT /opt/.
3. Run the custom controller	
	service citk8scontroller start

