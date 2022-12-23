package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	repomanagerv1alpha1 "github.com/pulp/pulp-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

type crd interface {
	convert(*kubernetes.Clientset) any
}

type pulp struct {
	ApiVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   metav1.ObjectMeta `json:"metadata"`
	Spec       AnsibleSpec       `json:"spec"`
	Status     any               `json:"status"`
}

type AnsibleSpec struct {
	AdminPasswordSecret          string                       `json:"admin_password_secret,omitempty"`
	Affinity                     *corev1.NodeAffinity         `json:"affinity,omitempty"`
	Api                          Api                          `json:"api,omitempty"`
	ContainerTokenSecret         string                       `json:"container_token_secret,omitempty"`
	Content                      Content                      `json:"content,omitempty"`
	DBFieldsEncryptionSecret     string                       `json:"db_fields_encryption_secret,omitempty"`
	DeploymentType               string                       `json:"deployment_type,omitempty"`
	FileStorageAccessMode        string                       `json:"file_storage_access_mode,omitempty"`
	FileStorageSize              string                       `json:"file_storage_size,omitempty"`
	FileStorageClass             string                       `json:"file_storage_storage_class,omitempty"`
	GunicornAPIWorkers           int                          `json:"gunicorn_api_workers,omitempty"`
	GunicornContentWorkers       int                          `json:"gunicorn_content_workers,omitempty"`
	GunicornTimeout              int                          `json:"gunicorn_timeout,omitempty"`
	HAProxyTimeout               string                       `json:"haproxy_timeout,omitempty"`
	Image                        string                       `json:"image,omitempty"`
	ImagePullPolicy              string                       `json:"image_pull_policy,omitempty"`
	ImagePullSecrets             []string                     `json:"image_pull_secrets,omitempty"`
	ImageVersion                 string                       `json:"image_version,omitempty"`
	ImageWeb                     string                       `json:"image_web,omitempty"`
	ImageWebVersion              string                       `json:"image_web_version,omitempty"`
	IngressAnnotations           map[string]string            `json:"ingress_annotations,omitempty"`
	IngressTLSSecret             string                       `json:"ingress_tls_secret,omitempty"`
	IngressType                  string                       `json:"ingress_type,omitempty"`
	NginxMaxBodySize             string                       `json:"nginx_client_max_body_size,omitempty"`
	NginxProxyConnectTimeout     string                       `json:"nginx_proxy_connect_timeout,omitempty"`
	NginxProxyReadTimeout        string                       `json:"nginx_proxy_read_timeout,omitempty"`
	NginxProxySendTimeout        string                       `json:"nginx_proxy_send_timeout,omitempty"`
	ObjectStorageAzureSecret     string                       `json:"object_storage_azure_secret,omitempty"`
	ObjectStorageS3Secret        string                       `json:"object_storage_s3_secret,omitempty"`
	PostgresDataPath             string                       `json:"postgres_data_path,omitempty"`
	PostgresExtraArgs            []string                     `json:"postgres_extra_args,omitempty"`
	PostgresHostAuthMethod       string                       `json:"postgres_host_auth_method,omitempty"`
	PostgresImage                string                       `json:"postgres_image,omitempty"`
	PostgresInitdbArgs           string                       `json:"postgres_initdb_args,omitempty"`
	PostgresResourceRequirements *corev1.ResourceRequirements `json:"postgres_resource_requirements,omitempty"`
	PostgresStorageClass         *string                      `json:"postgres_storage_class,omitempty"`
	PostgresStorageRequirements  string                       `json:"postgres_storage_requirements,omitempty"`
	PulpSettings                 runtime.RawExtension         `json:"pulp_settings,omitempty"`
	Redis                        Redis                        `json:"redis,omitempty"`
	RedisImage                   string                       `json:"redis_image,omitempty"`
	RedisResourceRequirements    corev1.ResourceRequirements  `json:"redis_resource_requirements,omitempty"`
	RedisStorageClass            string                       `json:"redis_storage_class,omitempty"`
	ResourceManager              ResourceManager              `json:"resource_manager,omitempty"`
	RouteHost                    string                       `json:"route_host,omitempty"`
	RouteTLSSecret               string                       `json:"route_tls_secret,omitempty"`
	SigningScriptsConfigmap      string                       `json:"signing_scripts_configmap,omitempty"`
	SigningSecret                string                       `json:"signing_secret,omitempty"`
	SSOSecret                    string                       `json:"sso_secret,omitempty"`
	StorageType                  string                       `json:"storage_type,omitempty"`
	Web                          Web                          `json:"web,omitempty"`
	Worker                       Web                          `json:"worker,omitempty"`

	// these are defined as string in ansible (but I'll let it the same way as we defined in go)
	Tolerations               []corev1.Toleration               `json:"tolerations,omitempty"`
	TopologySpreadConstraints []corev1.TopologySpreadConstraint `json:"topology_spread_constraints,omitempty"`

	// this is defined as map[string]string in golang
	NodeSelector string `json:"node_selector,omitempty"`

	// this is defined as int32 in golang
	NodePort string `json:"nodeport_port,omitempty"`

	// not found in golang
	Hostname                           string `json:"hostname,omitempty"`
	ImagePullSecret                    string `json:"image_pull_secret,omitempty"`
	LoadbalancerPort                   int    `json:"loadbalancer_port,omitempty"`
	LoadBalancerProtocol               string `json:"loadbalancer_protocol,omitempty"`
	NoLog                              string `json:"no_log,omitempty"`
	PostgresConfigurationSecret        string `json:"postgres_configuration_secret,omitempty"`
	PostgresKeepPVCAfterUpgrade        bool   `json:"postgres_keep_pvc_after_upgrade,omitempty"`
	PostgresLabelSelector              string `json:"postgres_label_selector,omitempty"`
	PostgresMigrantConfigurationSecret string `json:"postgres_migrant_configuration_secret,omitempty"`
	PostgresSelector                   string `json:"postgres_selector,omitempty"`
	PostgresToleration                 string `json:"postgres_tolerations,omitempty"`
	RouteTLSTerminationMechanism       string `json:"route_tls_termination_mechanism,omitempty"`
	ServiceAnnotations                 string `json:"service_annotations,omitempty"`
}

type Api struct {
	LogLevel             string                       `json:"log_level,omitempty"`
	Replicas             int32                        `json:"replicas,omitempty"`
	ResourceRequirements *corev1.ResourceRequirements `json:"resource_requirements,omitempty"`
	Strategy             *appsv1.DeploymentStrategy   `json:"strategy,omitempty"`
}
type Content struct {
	LogLevel             string                       `json:"log_level,omitempty"`
	Replicas             int32                        `json:"replicas,omitempty"`
	ResourceRequirements *corev1.ResourceRequirements `json:"resource_requirements,omitempty"`
	Strategy             *appsv1.DeploymentStrategy   `json:"strategy,omitempty"`
}

type Redis struct {
	LogLevel             string                       `json:"log_level,omitempty"`
	Replicas             int32                        `json:"replicas,omitempty"`
	ResourceRequirements *corev1.ResourceRequirements `json:"resource_requirements,omitempty"`
	Strategy             *appsv1.DeploymentStrategy   `json:"strategy,omitempty"`
}

type ResourceManager struct {
	Replicas             int32                        `json:"replicas,omitempty"`
	ResourceRequirements *corev1.ResourceRequirements `json:"resource_requirements,omitempty"`
	Strategy             *appsv1.DeploymentStrategy   `json:"strategy,omitempty"`
}

type Web struct {
	Replicas             int32                        `json:"replicas,omitempty"`
	ResourceRequirements *corev1.ResourceRequirements `json:"resource_requirements,omitempty"`
	Strategy             *appsv1.DeploymentStrategy   `json:"strategy,omitempty"`
}

type Worker struct {
	Replicas             int32                        `json:"replicas,omitempty"`
	ResourceRequirements *corev1.ResourceRequirements `json:"resource_requirements,omitempty"`
	Strategy             *appsv1.DeploymentStrategy   `json:"strategy,omitempty"`
}

func (ansiblePulp pulp) convert(clientset *kubernetes.Clientset) {
	api := "/apis/pulp.pulpproject.org/v1beta1"
	namespace := "pulp"
	resource := "pulps"
	resourceName := "example-pulp"

	goApi := "repo-manager.pulpproject.org/v1alpha1"
	goKind := "Pulp"

	ctx := context.TODO()

	data, err := clientset.RESTClient().
		Get().
		AbsPath(api).
		Namespace(namespace).
		Resource(resource).
		Name(resourceName).
		DoRaw(ctx)

	if err != nil {
		fmt.Println(err)
		return
	}

	json.Unmarshal(data, &ansiblePulp)
	//fmt.Println(ansiblePulp.ApiVersion)
	//fmt.Println(ansiblePulp.Kind)
	//fmt.Println(ansiblePulp.Metadata)
	fmt.Println(ansiblePulp.Spec)

	ansibleCRDValues := reflect.ValueOf(ansiblePulp.Spec)
	ansibleCRDTypes := ansibleCRDValues.Type()
	for i := 0; i < ansibleCRDValues.NumField(); i++ {
		fmt.Printf("%v: %v\n", ansibleCRDTypes.Field(i).Name, ansibleCRDValues.Field(i))
	}

	pulpNew := &repomanagerv1alpha1.Pulp{
		TypeMeta: metav1.TypeMeta{
			APIVersion: goApi,
			Kind:       goKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ansiblePulp.Metadata.Name,
			Namespace: ansiblePulp.Metadata.Namespace,
		},
		Spec: repomanagerv1alpha1.PulpSpec{
			DeploymentType:           ansiblePulp.Spec.DeploymentType,
			FileStorageSize:          ansiblePulp.Spec.FileStorageSize,
			FileStorageAccessMode:    ansiblePulp.Spec.FileStorageAccessMode,
			FileStorageClass:         ansiblePulp.Spec.FileStorageClass,
			ObjectStorageAzureSecret: ansiblePulp.Spec.ObjectStorageAzureSecret,
			ObjectStorageS3Secret:    ansiblePulp.Spec.ObjectStorageS3Secret,
			DBFieldsEncryptionSecret: ansiblePulp.Spec.DBFieldsEncryptionSecret,
			SigningSecret:            ansiblePulp.Spec.SigningSecret,
			SigningScriptsConfigmap:  ansiblePulp.Spec.SigningScriptsConfigmap,
			StorageType:              ansiblePulp.Spec.StorageType,
			IngressType:              ansiblePulp.Spec.IngressType,
			IngressAnnotations:       ansiblePulp.Spec.IngressAnnotations,
			IngressTLSSecret:         ansiblePulp.Spec.IngressTLSSecret,
			RouteHost:                ansiblePulp.Spec.RouteHost,
			RouteTLSSecret:           ansiblePulp.Spec.RouteTLSSecret,
			HAProxyTimeout:           ansiblePulp.Spec.HAProxyTimeout,
			NginxMaxBodySize:         ansiblePulp.Spec.NginxMaxBodySize,
			NginxProxyBodySize:       ansiblePulp.Spec.NginxMaxBodySize,
			NginxProxyReadTimeout:    ansiblePulp.Spec.NginxProxyReadTimeout,
			NginxProxyConnectTimeout: ansiblePulp.Spec.NginxProxyConnectTimeout,
			NginxProxySendTimeout:    ansiblePulp.Spec.NginxProxySendTimeout,
			ContainerTokenSecret:     ansiblePulp.Spec.ContainerTokenSecret,
			Image:                    ansiblePulp.Spec.Image,
			ImageVersion:             ansiblePulp.Spec.ImageVersion,
			ImagePullPolicy:          ansiblePulp.Spec.ImagePullPolicy,
			PulpSettings:             ansiblePulp.Spec.PulpSettings,
			ImageWeb:                 ansiblePulp.Spec.ImageWeb,
			ImageWebVersion:          ansiblePulp.Spec.ImageWebVersion,
			AdminPasswordSecret:      ansiblePulp.Spec.AdminPasswordSecret,
			ImagePullSecrets:         ansiblePulp.Spec.ImagePullSecrets,
			SSOSecret:                ansiblePulp.Spec.SSOSecret,
			Api: repomanagerv1alpha1.Api{
				Replicas:                  ansiblePulp.Spec.Api.Replicas,
				Tolerations:               ansiblePulp.Spec.Tolerations,
				TopologySpreadConstraints: ansiblePulp.Spec.TopologySpreadConstraints,
				GunicornTimeout:           ansiblePulp.Spec.GunicornTimeout,
				GunicornWorkers:           ansiblePulp.Spec.GunicornAPIWorkers,
				//ResourceRequirements:      *ansiblePulp.Spec.Api.ResourceRequirements,
				ReadinessProbe: nil,
				LivenessProbe:  nil,
				PDB:            nil,
				//Strategy:       *ansiblePulp.Spec.Api.Strategy,
				//Affinity: ansiblePulp.Spec.Affinity,
				//NodeSelector: ansiblePulp.Spec.NodeSelector,
			},
			Content: repomanagerv1alpha1.Content{
				Replicas:                  ansiblePulp.Spec.Content.Replicas,
				Tolerations:               ansiblePulp.Spec.Tolerations,
				TopologySpreadConstraints: ansiblePulp.Spec.TopologySpreadConstraints,
				GunicornTimeout:           ansiblePulp.Spec.GunicornTimeout,
				GunicornWorkers:           ansiblePulp.Spec.GunicornContentWorkers,
				ResourceRequirements:      *ansiblePulp.Spec.Content.ResourceRequirements,
				ReadinessProbe:            nil,
				LivenessProbe:             nil,
				PDB:                       nil,
				Strategy:                  *ansiblePulp.Spec.Content.Strategy,
				//Affinity: ansiblePulp.Spec.Affinity,
				//NodeSelector: ansiblePulp.Spec.NodeSelector,
			},
			Worker: repomanagerv1alpha1.Worker{
				Replicas:                  ansiblePulp.Spec.Worker.Replicas,
				Tolerations:               ansiblePulp.Spec.Tolerations,
				TopologySpreadConstraints: ansiblePulp.Spec.TopologySpreadConstraints,
				ResourceRequirements:      *ansiblePulp.Spec.Worker.ResourceRequirements,
				ReadinessProbe:            nil,
				LivenessProbe:             nil,
				PDB:                       nil,
				Strategy:                  *ansiblePulp.Spec.Worker.Strategy,
				//Affinity: ansiblePulp.Spec.Affinity,
				//NodeSelector: ansiblePulp.Spec.NodeSelector,
			},
			Web: repomanagerv1alpha1.Web{
				Replicas:             ansiblePulp.Spec.Web.Replicas,
				ResourceRequirements: *ansiblePulp.Spec.Worker.ResourceRequirements,
				ReadinessProbe:       nil,
				LivenessProbe:        nil,
				PDB:                  nil,
				//Affinity: ansiblePulp.Spec.Affinity,
				//NodeSelector: ansiblePulp.Spec.NodeSelector,
			},
			Database: repomanagerv1alpha1.Database{
				Affinity:                    nil,
				PostgresImage:               ansiblePulp.Spec.PostgresImage,
				PostgresExtraArgs:           ansiblePulp.Spec.PostgresExtraArgs,
				PostgresDataPath:            ansiblePulp.Spec.PostgresDataPath,
				PostgresInitdbArgs:          ansiblePulp.Spec.PostgresInitdbArgs,
				PostgresHostAuthMethod:      ansiblePulp.Spec.PostgresHostAuthMethod,
				ResourceRequirements:        *ansiblePulp.Spec.PostgresResourceRequirements,
				PostgresStorageRequirements: ansiblePulp.Spec.PostgresStorageRequirements,
				PostgresStorageClass:        ansiblePulp.Spec.PostgresStorageClass,
				ReadinessProbe:              nil,
				LivenessProbe:               nil,
				//ExternalDBSecret: "",
				//PostgresVersion: "",
				//PostgresPort: 5432,
				//PostgresSSLMode: "prefer",
				//NodeSelector:           ansiblePulp.Spec.PostgresSelector,
				//Tolerations: ansiblePulp.Spec.PostgresToleration,
				//PVC:          "",
			},
			Cache: repomanagerv1alpha1.Cache{
				RedisImage:                ansiblePulp.Spec.RedisImage,
				RedisStorageClass:         ansiblePulp.Spec.RedisStorageClass,
				RedisResourceRequirements: ansiblePulp.Spec.RedisResourceRequirements,
				ReadinessProbe:            nil,
				LivenessProbe:             nil,
				Affinity:                  nil,
				Tolerations:               nil,
				NodeSelector:              nil,
				Strategy:                  *ansiblePulp.Spec.Redis.Strategy,
				//ExternalCacheSecret: "",
				//Enabled: true,
				//RedisPort: 6379,
				//PVC: "",
			},
		},
	}
	body, err := json.Marshal(pulpNew)
	if err != nil {
		fmt.Println(err)
		return
	}

	data, err = clientset.RESTClient().
		Post().
		AbsPath("/apis/" + goApi + "/namespaces/" + ansiblePulp.Metadata.Namespace + "/" + goKind).
		Body(body).
		DoRaw(context.TODO())

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(data)
}

func main() {
	config := ctrl.GetConfigOrDie()
	clientset := kubernetes.NewForConfigOrDie(config)

	ansiblePulp := pulp{}
	ansiblePulp.convert(clientset)
}
