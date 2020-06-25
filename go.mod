module fate-operator

go 1.13

require (
	github.com/FederatedAI/KubeFATE/k8s-deploy v0.0.0-20200413035739-b2e0b5479815
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-logr/logr v0.1.0
	github.com/json-iterator/go v1.1.9
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	golang.org/x/net v0.0.0-20191028085509-fe3aa8a45271
	gopkg.in/ffmt.v1 v1.5.6
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.0.0-20191219154910-1528d4eea6dd
	sigs.k8s.io/controller-runtime v0.5.2
	sigs.k8s.io/kustomize/kustomize/v3 v3.5.4 // indirect
	sigs.k8s.io/yaml v1.1.0
)

replace (
	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.9.0

	//github.com/FederatedAI/KubeFATE/k8s-deploy v0.0.0-20200413035739-b2e0b5479815 => github.com/FederatedAI/KubeFATE/k8s-deploy v0.0.0
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	// Kubernetes imports github.com/miekg/dns at a newer version but it is used
	// by a package Helm does not need. Go modules resolves all packages rather
	// than just those in use (like Glide and dep do). This sets the version
	// to the one oras needs. If oras is updated the version should be updated
	// as well.
	github.com/miekg/dns => github.com/miekg/dns v0.0.0-20181005163659-0d29b283ac0f
	gopkg.in/inf.v0 v0.9.1 => github.com/go-inf/inf v0.9.1
	gopkg.in/square/go-jose.v2 v2.3.0 => github.com/square/go-jose v2.3.0+incompatible

	rsc.io/letsencrypt => github.com/dmcgowan/letsencrypt v0.0.0-20160928181947-1847a81d2087
)
