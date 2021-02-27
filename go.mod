module github.com/argoproj/pkg

go 1.14

require (
	github.com/aws/aws-sdk-go v1.33.16
	github.com/dustin/go-humanize v1.0.0
	github.com/felixge/httpsnoop v1.0.1
	github.com/golang/protobuf v1.3.3
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/minio/minio-go/v7 v7.0.2
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.6.1
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	k8s.io/api v0.17.8
	k8s.io/apimachinery v0.17.8
	k8s.io/client-go v0.17.8
	k8s.io/klog/v2 v2.5.0
)

replace (
	// https://github.com/kubernetes/kubernetes/issues/79384#issuecomment-505627280
	k8s.io/api => k8s.io/api v0.17.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.8 // indirect
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.8 // indirect
	k8s.io/apiserver => k8s.io/apiserver v0.17.8
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.8
	k8s.io/client-go => k8s.io/client-go v0.17.8
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.8
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.8
	k8s.io/code-generator => k8s.io/code-generator v0.17.8
	k8s.io/component-base => k8s.io/component-base v0.17.8
	k8s.io/cri-api => k8s.io/cri-api v0.17.8
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.8
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.8
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.8
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.8
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.8
	k8s.io/kubectl => k8s.io/kubectl v0.17.8
	k8s.io/kubelet => k8s.io/kubelet v0.17.8
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.8
	k8s.io/metrics => k8s.io/metrics v0.17.8
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.8
)
