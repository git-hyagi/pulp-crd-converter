package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	repomanagerv1alpha1 "github.com/pulp/pulp-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	// ansible subscription data
	oldSubscriptionName      string
	oldSubscriptionNamespace string

	// go subscription data
	newSubscriptionNamespace           string
	newSubscriptionName                string
	newSubscriptionChannel             string
	newSubscriptionInstallPlanApproval string
	newSubscriptionSource              string
	newSubscriptionSourceNamespace     string
	newSubscriptionStartingCSV         string

	// CRD data
	oldApi          string
	oldResource     string
	oldResourceName string
	newApi          string
	newKind         string
	newResourceName string
	newResource     string
	oldDBPVC        string
	oldDBSVC        string
	oldDBSts        string
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

func (pulp *pulp) getCurrentDBPVC(clientset *kubernetes.Clientset) error {
	fmt.Println("🔎 Retrieving the current Database PVC ...")
	data, err := clientset.RESTClient().
		Get().
		AbsPath("/api/v1").
		Namespace(pulp.oldSubscriptionNamespace).
		Resource("persistentvolumeclaims").
		Param("labelSelector", "app.kubernetes.io/component=database,app.kubernetes.io/managed-by="+pulp.oldSubscriptionName).
		DoRaw(context.TODO())
	if err != nil {
		fmt.Println("❌ Failed to find Database PVC:", err)
		return err
	}
	pvcList := &corev1.PersistentVolumeClaimList{}
	json.Unmarshal(data, pvcList)
	if len(pvcList.Items) >= 1 {
		pulp.oldDBPVC = pvcList.Items[0].ObjectMeta.Name
		fmt.Println("Migrator will use the following PVC to the database pods:", pvcList.Items[0].ObjectMeta.Name)
	}

	return nil
}

func (pulp *pulp) getCurrentDBService(clientset *kubernetes.Clientset) error {
	fmt.Println("🔎 Retrieving the current Database Service ...")
	data, err := clientset.RESTClient().
		Get().
		AbsPath("/api/v1").
		Namespace(pulp.oldSubscriptionNamespace).
		Resource("services").
		Param("labelSelector", "app.kubernetes.io/component=database,app.kubernetes.io/managed-by="+pulp.oldSubscriptionName).
		DoRaw(context.TODO())
	if err != nil {
		fmt.Println("❌ Failed to find Database Service:", err)
		return err
	}
	svcList := &corev1.ServiceList{}
	json.Unmarshal(data, svcList)
	if len(svcList.Items) >= 1 {
		pulp.oldDBSVC = svcList.Items[0].ObjectMeta.Name
		fmt.Println("Migrator will use the following SVC to the database pods:", svcList.Items[0].ObjectMeta.Name)
	} else {
		fmt.Println("❌ Failed to find Database Service")
		return fmt.Errorf("database Service not found")
	}

	return nil
}

func (pulp *pulp) getCurrentDBSts(clientset *kubernetes.Clientset) error {
	fmt.Println("🔎 Retrieving the current Database StatefulSet ...")
	data, err := clientset.RESTClient().
		Get().
		AbsPath("/apis/apps/v1").
		Namespace(pulp.oldSubscriptionNamespace).
		Resource("statefulsets").
		Param("labelSelector", "app.kubernetes.io/component=database,app.kubernetes.io/managed-by="+pulp.oldSubscriptionName).
		DoRaw(context.TODO())
	if err != nil {
		fmt.Println("❌ Failed to find Database Service:", err)
		return err
	}
	stsList := &appsv1.StatefulSetList{}
	json.Unmarshal(data, stsList)
	if len(stsList.Items) >= 1 {
		pulp.oldDBSts = stsList.Items[0].ObjectMeta.Name
		fmt.Println("Migrator will downscale the following StatefulSet to 0 replica pods:", stsList.Items[0].ObjectMeta.Name)
	} else {
		fmt.Println("❌ Failed to find Database StatefulSet")
		return fmt.Errorf("database StatefulSet not found")
	}

	return nil
}

func (pulp pulp) getCurrentCSV(clientset *kubernetes.Clientset) (string, error) {
	fmt.Println("🔎 Retrieving the current csv from subscription", pulp.oldSubscriptionName, "...")
	data, err := clientset.RESTClient().
		Get().
		AbsPath("/apis/operators.coreos.com/v1alpha1").
		Namespace(pulp.oldSubscriptionNamespace).
		Resource("subscriptions").
		Name(pulp.oldSubscriptionName).
		DoRaw(context.TODO())
	if err != nil {
		fmt.Println("❌ Failed to find Subscription:", err)
		return "", err
	}
	sub := &operatorsv1alpha1.Subscription{}
	json.Unmarshal(data, sub)
	currentCSV := sub.Status.CurrentCSV
	fmt.Println("Current CSV Name:", currentCSV)
	return currentCSV, nil
}

func (pulp pulp) deleteSubscription(clientset *kubernetes.Clientset) error {
	fmt.Println("🗑️  Deleting", pulp.oldSubscriptionName, "subscription ...")
	data, err := clientset.RESTClient().
		Delete().
		AbsPath("/apis/operators.coreos.com/v1alpha1").
		Namespace(pulp.oldSubscriptionNamespace).
		Resource("subscriptions").
		Name(pulp.oldSubscriptionName).
		DoRaw(context.TODO())
	if err != nil {
		fmt.Println("❌ Failed to find Subscription:", err)
		return err
	}

	fmt.Println(string(data))
	return nil
}

func (pulp pulp) deleteCSV(clientset *kubernetes.Clientset, csvName string) error {
	fmt.Println("🗑️  Deleting", csvName, "CSV ...")
	data, err := clientset.RESTClient().
		Delete().
		AbsPath("/apis/operators.coreos.com/v1alpha1").
		Namespace(pulp.oldSubscriptionNamespace).
		Resource("clusterserviceversions").
		Name(csvName).
		DoRaw(context.TODO())
	if err != nil {
		fmt.Println("❌ Failed to find Subscription:", err)
		return err
	}

	fmt.Println(string(data))
	return nil
}

func (pulp *pulp) updateDBService(clientset *kubernetes.Clientset) error {
	fmt.Println("Updating " + pulp.oldDBSVC + " Database Service ...")

	// remove old label selectors
	labels := []string{"app.kubernetes.io/instance", "app.kubernetes.io/component", "app.kubernetes.io/managed-by", "app.kubernetes.io/name", "app.kubernetes.io/part-of", "app.kubernetes.io/version"}
	for _, label := range labels {
		_, err := clientset.RESTClient().
			Patch(types.MergePatchType).
			AbsPath("/api/v1").
			Namespace(pulp.oldSubscriptionNamespace).
			Resource("services").
			Name(pulp.oldDBSVC).
			Param("fieldManager", "kubectl-label").
			Body([]byte(`{"spec":{"selector":{"` + label + `":null}}}`)).
			DoRaw(context.TODO())
		if err != nil {
			fmt.Println("❌ Failed to remove old labels from Database Service:", err)
			return err
		}
	}

	// configure the selector for the new DB pods
	newLabels := map[string]string{
		"app":     "postgresql",
		"pulp_cr": pulp.newResourceName,
	}
	for k, v := range newLabels {
		_, err := clientset.RESTClient().
			Patch(types.MergePatchType).
			AbsPath("/api/v1").
			Namespace(pulp.oldSubscriptionNamespace).
			Resource("services").
			Name(pulp.oldDBSVC).
			Param("fieldManager", "kubectl-label").
			Body([]byte(`{"spec":{"selector":{"` + k + `":"` + v + `"}}}`)).
			DoRaw(context.TODO())
		if err != nil {
			fmt.Println("❌ Failed to add new labels to the Database Service:", err)
			return err
		}
	}
	return nil
}

func (pulp pulp) subscribe(clientset *kubernetes.Clientset) error {
	fmt.Println("Subscribing to the new Operator version ...")
	newSubscription := &operatorsv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operators.coreos.com/v1alpha1",
			Kind:       "Subscription",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pulp.newSubscriptionName,
			Namespace: pulp.newSubscriptionNamespace,
		},
		Spec: &operatorsv1alpha1.SubscriptionSpec{
			Channel:                pulp.newSubscriptionChannel,
			InstallPlanApproval:    operatorsv1alpha1.Approval(pulp.newSubscriptionInstallPlanApproval),
			CatalogSource:          pulp.newSubscriptionSource,
			CatalogSourceNamespace: pulp.newSubscriptionSourceNamespace,
			StartingCSV:            pulp.newSubscriptionStartingCSV,
			Package:                pulp.newSubscriptionName,
		},
	}
	body, err := json.Marshal(newSubscription)
	if err != nil {
		fmt.Println("❌ Failed to serialize new Pulp CR:", err)
		return err
	}
	fmt.Println(string(body))

	data, err := clientset.RESTClient().
		Post().
		AbsPath("/apis/operators.coreos.com/v1alpha1").
		Namespace(pulp.newSubscriptionNamespace).
		Resource("subscriptions").
		Name(pulp.newSubscriptionName).
		Body(body).
		DoRaw(context.TODO())
	if err != nil {
		fmt.Println("❌ Failed to create Subscription:", err)
		return err
	}

	fmt.Println(string(data))
	return nil
}

func (pulp pulp) deleteDeployments(clientset *kubernetes.Clientset) error {
	components := []string{"api", "content-server", "worker", "webserver", "cache"}

	for _, component := range components {
		_, err := clientset.RESTClient().
			Delete().
			AbsPath("/apis/apps/v1").
			Namespace(pulp.oldSubscriptionNamespace).
			Resource("deployments").
			Param("labelSelector", "app.kubernetes.io/component="+component).
			DoRaw(context.TODO())
		if err != nil {
			fmt.Println("❌ Failed to find", component, "deployment:", err)
			return err
		} else {
			fmt.Println("🗑️  Deleting", component, "deployment ...")
		}
	}
	return nil
}

func (pulp pulp) downscaleDBReplicas(clientset *kubernetes.Clientset) error {
	fmt.Println("Scaling old Database STS to 0 replicas ...")
	if _, err := clientset.RESTClient().
		Patch(types.MergePatchType).
		AbsPath("/apis/apps/v1").
		Namespace(pulp.oldSubscriptionNamespace).
		Resource("statefulsets").
		Name(pulp.oldDBSts).
		Suffix("scale").
		Body([]byte(`{"spec":{"replicas":0}}`)).
		DoRaw(context.TODO()); err != nil {
		fmt.Println("❌ Failed to set "+pulp.oldDBSts+" STS to 0 replicas:", err)
		return err
	}
	return nil
}

func (pulp pulp) convert(clientset *kubernetes.Clientset) error {
	ctx := context.TODO()

	fmt.Println("Converting Pulp CR to the new CRD ...")
	data, err := clientset.RESTClient().
		Get().
		AbsPath(pulp.oldApi).
		Namespace(pulp.oldSubscriptionNamespace).
		Resource(pulp.oldResource).
		Name(pulp.oldResourceName).
		DoRaw(ctx)

	if err != nil {
		fmt.Println("❌ Failed to find old Pulp CR:", err)
		return err
	}

	json.Unmarshal(data, &pulp)

	apiResources := corev1.ResourceRequirements{}
	if pulp.Spec.Api.ResourceRequirements != nil {
		apiResources = *pulp.Spec.Api.ResourceRequirements
	}
	contentResources := corev1.ResourceRequirements{}
	if pulp.Spec.Content.ResourceRequirements != nil {
		contentResources = *pulp.Spec.Content.ResourceRequirements
	}
	workerResources := corev1.ResourceRequirements{}
	if pulp.Spec.Worker.ResourceRequirements != nil {
		workerResources = *pulp.Spec.Worker.ResourceRequirements
	}
	webResources := corev1.ResourceRequirements{}
	if pulp.Spec.Web.ResourceRequirements != nil {
		webResources = *pulp.Spec.Web.ResourceRequirements
	}
	dbResources := corev1.ResourceRequirements{}
	if pulp.Spec.PostgresResourceRequirements != nil {
		dbResources = *pulp.Spec.PostgresResourceRequirements
	}

	apiStrategy := appsv1.DeploymentStrategy{}
	if pulp.Spec.Api.Strategy != nil {
		apiStrategy = *pulp.Spec.Api.Strategy
	}
	contentStrategy := appsv1.DeploymentStrategy{}
	if pulp.Spec.Content.Strategy != nil {
		contentStrategy = *pulp.Spec.Content.Strategy
	}
	workerStrategy := appsv1.DeploymentStrategy{}
	if pulp.Spec.Worker.Strategy != nil {
		workerStrategy = *pulp.Spec.Worker.Strategy
	}
	cacheStrategy := appsv1.DeploymentStrategy{}
	if pulp.Spec.Web.Strategy != nil {
		cacheStrategy = *pulp.Spec.Redis.Strategy
	}

	imagePullSecrets := pulp.Spec.ImagePullSecrets
	if pulp.Spec.ImagePullSecret != "" {
		imagePullSecrets = append(imagePullSecrets, pulp.Spec.ImagePullSecret)
	}

	pulpPVC := ""
	if len(pulp.Spec.ObjectStorageAzureSecret) == 0 && len(pulp.Spec.ObjectStorageS3Secret) == 0 {
		pulpPVC = pulp.oldResourceName + "-file-storage"
	}
	redisPVC := pulp.oldResourceName + "-redis-data"

	// Defining file_storage_class as "" to avoid conflict with pvc definition.
	// In go version we are verifying multiple storage definitions,
	// in ansible, when none of s3 or azure blob secrets are provided, the operator
	// will provision a PVC. If a SC is provided, it will define the PVC spec with it,
	// if not, no SC will be defined and k8s will try to use an available PV that fits
	// the spec of the PVC.
	fileStorageClass := ""
	cacheStorageClass := ""
	dbStorageClass := (*string)(nil)

	deploymentType := "pulp"
	if isGalaxy, _ := regexp.MatchString(".*galaxy.*", pulp.Spec.Image); isGalaxy {
		deploymentType = "galaxy"
	}

	pulpNew := &repomanagerv1alpha1.Pulp{
		TypeMeta: metav1.TypeMeta{
			APIVersion: pulp.newApi,
			Kind:       pulp.newKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pulp.newResourceName,
			Namespace: pulp.newSubscriptionNamespace,
		},
		Spec: repomanagerv1alpha1.PulpSpec{
			DeploymentType:           deploymentType,
			FileStorageSize:          pulp.Spec.FileStorageSize,
			FileStorageAccessMode:    pulp.Spec.FileStorageAccessMode,
			FileStorageClass:         fileStorageClass,
			PVC:                      pulpPVC,
			ObjectStorageAzureSecret: pulp.Spec.ObjectStorageAzureSecret,
			ObjectStorageS3Secret:    pulp.Spec.ObjectStorageS3Secret,
			DBFieldsEncryptionSecret: pulp.Spec.DBFieldsEncryptionSecret,
			SigningSecret:            pulp.Spec.SigningSecret,
			SigningScriptsConfigmap:  pulp.Spec.SigningScriptsConfigmap,
			StorageType:              pulp.Spec.StorageType,
			IngressType:              pulp.Spec.IngressType,
			IngressAnnotations:       pulp.Spec.IngressAnnotations,
			IngressTLSSecret:         pulp.Spec.IngressTLSSecret,
			RouteHost:                pulp.Spec.RouteHost,
			RouteTLSSecret:           pulp.Spec.RouteTLSSecret,
			HAProxyTimeout:           pulp.Spec.HAProxyTimeout,
			NginxMaxBodySize:         pulp.Spec.NginxMaxBodySize,
			NginxProxyBodySize:       pulp.Spec.NginxMaxBodySize,
			NginxProxyReadTimeout:    pulp.Spec.NginxProxyReadTimeout,
			NginxProxyConnectTimeout: pulp.Spec.NginxProxyConnectTimeout,
			NginxProxySendTimeout:    pulp.Spec.NginxProxySendTimeout,
			ContainerTokenSecret:     pulp.Spec.ContainerTokenSecret,
			Image:                    pulp.Spec.Image,
			ImageVersion:             pulp.Spec.ImageVersion,
			ImagePullPolicy:          pulp.Spec.ImagePullPolicy,
			PulpSettings:             pulp.Spec.PulpSettings,
			ImageWeb:                 pulp.Spec.ImageWeb,
			ImageWebVersion:          pulp.Spec.ImageWebVersion,
			AdminPasswordSecret:      pulp.Spec.AdminPasswordSecret,
			ImagePullSecrets:         imagePullSecrets,
			SSOSecret:                pulp.Spec.SSOSecret,
			Api: repomanagerv1alpha1.Api{
				Replicas:                  pulp.Spec.Api.Replicas,
				Tolerations:               pulp.Spec.Tolerations,
				TopologySpreadConstraints: pulp.Spec.TopologySpreadConstraints,
				GunicornTimeout:           pulp.Spec.GunicornTimeout,
				GunicornWorkers:           pulp.Spec.GunicornAPIWorkers,
				ResourceRequirements:      apiResources,
				ReadinessProbe:            nil,
				LivenessProbe:             nil,
				PDB:                       nil,
				Strategy:                  apiStrategy,
				//Affinity: pulp.Spec.Affinity,
				//NodeSelector: pulp.Spec.NodeSelector,
			},
			Content: repomanagerv1alpha1.Content{
				Replicas:                  pulp.Spec.Content.Replicas,
				Tolerations:               pulp.Spec.Tolerations,
				TopologySpreadConstraints: pulp.Spec.TopologySpreadConstraints,
				GunicornTimeout:           pulp.Spec.GunicornTimeout,
				GunicornWorkers:           pulp.Spec.GunicornContentWorkers,
				ResourceRequirements:      contentResources,
				ReadinessProbe:            nil,
				LivenessProbe:             nil,
				PDB:                       nil,
				Strategy:                  contentStrategy,
				//Affinity: pulp.Spec.Affinity,
				//NodeSelector: pulp.Spec.NodeSelector,
			},
			Worker: repomanagerv1alpha1.Worker{
				Replicas:                  pulp.Spec.Worker.Replicas,
				Tolerations:               pulp.Spec.Tolerations,
				TopologySpreadConstraints: pulp.Spec.TopologySpreadConstraints,
				ResourceRequirements:      workerResources,
				ReadinessProbe:            nil,
				LivenessProbe:             nil,
				PDB:                       nil,
				Strategy:                  workerStrategy,
				//Affinity: pulp.Spec.Affinity,
				//NodeSelector: pulp.Spec.NodeSelector,
			},
			Web: repomanagerv1alpha1.Web{
				Replicas:             pulp.Spec.Web.Replicas,
				ResourceRequirements: webResources,
				ReadinessProbe:       nil,
				LivenessProbe:        nil,
				PDB:                  nil,
				//Affinity: pulp.Spec.Affinity,
				//NodeSelector: pulp.Spec.NodeSelector,
			},
			Database: repomanagerv1alpha1.Database{
				Affinity:                    nil,
				PostgresImage:               pulp.Spec.PostgresImage,
				PostgresExtraArgs:           pulp.Spec.PostgresExtraArgs,
				PostgresDataPath:            pulp.Spec.PostgresDataPath,
				PostgresInitdbArgs:          pulp.Spec.PostgresInitdbArgs,
				PostgresHostAuthMethod:      pulp.Spec.PostgresHostAuthMethod,
				ResourceRequirements:        dbResources,
				PostgresStorageRequirements: pulp.Spec.PostgresStorageRequirements,
				PostgresStorageClass:        dbStorageClass,
				ReadinessProbe:              nil,
				LivenessProbe:               nil,
				PVC:                         pulp.oldDBPVC,
				//ExternalDBSecret: "",
				//PostgresVersion: "",
				//PostgresPort: 5432,
				//PostgresSSLMode: "prefer",
				//NodeSelector:           pulp.Spec.PostgresSelector,
				//Tolerations: pulp.Spec.PostgresToleration,
			},
			Cache: repomanagerv1alpha1.Cache{
				RedisImage:                pulp.Spec.RedisImage,
				RedisStorageClass:         cacheStorageClass,
				RedisResourceRequirements: pulp.Spec.RedisResourceRequirements,
				ReadinessProbe:            nil,
				LivenessProbe:             nil,
				Affinity:                  nil,
				Tolerations:               nil,
				NodeSelector:              nil,
				Strategy:                  cacheStrategy,
				PVC:                       redisPVC,
				//ExternalCacheSecret: "",
				//Enabled: true,
				//RedisPort: 6379,
			},
		},
	}
	body, err := json.Marshal(pulpNew)
	if err != nil {
		fmt.Println("❌ Failed to serialize new Pulp CR:", err)
		return err
	}

	fmt.Println("Create new CR:", string(body))

	retries := 10
	tried := 0
	for ; tried < retries; tried++ {
		if data, err = clientset.RESTClient().
			Get().
			AbsPath("/apis/" + pulp.newApi).
			DoRaw(ctx); err != nil {
			fmt.Println("Waiting for new CRD be created ... :", err)
			time.Sleep(time.Second * 5)
		} else {
			fmt.Println("CRD:", string(data))
			break
		}
	}

	if tried == 10 {
		fmt.Println("❌ ERROR! Golang CRD not found!")
		return err
	}

	_, err = clientset.RESTClient().
		Post().
		AbsPath("/apis/" + pulp.newApi + "/namespaces/" + pulp.newSubscriptionNamespace + "/" + pulp.newResource).
		Body(body).
		DoRaw(ctx)

	if err != nil {
		fmt.Println("❌ Failed to create new Pulp CR:", err)
		return err
	}
	return nil
}

func main() {
	config := ctrl.GetConfigOrDie()
	clientset := kubernetes.NewForConfigOrDie(config)

	// required variables
	namespace := os.Getenv("PULP_NAMESPACE")
	if namespace == "" {
		fmt.Println("Missing definition of PULP_NAMESPACE env var!")
		return
	}
	oldResourceName := os.Getenv("PULP_RESOURCE_NAME")
	if oldResourceName == "" {
		fmt.Println("Missing definition of PULP_RESOURCE_NAME")
		return
	}
	newResourceName := os.Getenv("NEW_PULP_RESOURCE_NAME")
	if newResourceName == "" {
		newResourceName = oldResourceName
	}

	// variables default values
	oldSubscriptionName := os.Getenv("PULP_SUBSCRIPTION_NAME")
	if oldSubscriptionName == "" {
		oldSubscriptionName = "pulp-operator"
	}
	newSubscriptionName := os.Getenv("NEW_PULP_SUBSCRIPTION_NAME")
	if newSubscriptionName == "" {
		newSubscriptionName = oldSubscriptionName
	}
	newSubscriptionChannel := os.Getenv("NEW_SUBSCRIPTION_CHANNEL")
	if newSubscriptionChannel == "" {
		newSubscriptionChannel = "beta"
	}
	newSubscriptionInstallPlanApproval := os.Getenv("NEW_SUBSCRIPTION_INSTALL_PLAN_APPROVAL")
	if newSubscriptionInstallPlanApproval == "" {
		newSubscriptionInstallPlanApproval = "Automatic"
	}
	newSubscriptionSource := os.Getenv("NEW_SUBSCRIPTION_SOURCE")
	if newSubscriptionSource == "" {
		newSubscriptionSource = "community-operators"
	}
	newSubscriptionSourceNamespace := os.Getenv("NEW_SUBSCRIPTION_SOURCE_NAMESPACE")
	if newSubscriptionSourceNamespace == "" {
		newSubscriptionSourceNamespace = "openshift-marketplace"
	}
	newSubscriptionStartingCSV := os.Getenv("NEW_SUBSCRIPTION_STARTING_CSV")
	if newSubscriptionStartingCSV == "" {
		newSubscriptionStartingCSV = "pulp-operator.v1.0.0-alpha.4"
	}
	newApi := os.Getenv("NEW_PULP_API")
	if newApi == "" {
		newApi = "repo-manager.pulpproject.org/v1alpha1"
	}
	oldApi := os.Getenv("PULP_API")
	if oldApi == "" {
		oldApi = "/apis/pulp.pulpproject.org/v1beta1"
	}
	newKind := os.Getenv("NEW_PULP_KIND")
	if newKind == "" {
		newKind = "Pulp"
	}
	oldResource := os.Getenv("PULP_RESOURCE")
	if oldResource == "" {
		oldResource = "pulps"
	}
	newResource := os.Getenv("NEW_PULP_RESOURCE")
	if newResource == "" {
		newResource = oldResource
	}

	ansiblePulp := pulp{
		oldSubscriptionName:                oldSubscriptionName,
		oldSubscriptionNamespace:           namespace,
		newSubscriptionNamespace:           namespace,
		newSubscriptionName:                newSubscriptionName,
		newSubscriptionChannel:             newSubscriptionChannel,
		newSubscriptionInstallPlanApproval: newSubscriptionInstallPlanApproval,
		newSubscriptionSource:              newSubscriptionSource,
		newSubscriptionSourceNamespace:     newSubscriptionSourceNamespace,
		newSubscriptionStartingCSV:         newSubscriptionStartingCSV,
		newApi:                             newApi,
		newKind:                            newKind,
		newResourceName:                    newResourceName,
		newResource:                        newResource,
		oldApi:                             oldApi,
		oldResource:                        oldResource,
		oldResourceName:                    oldResourceName,
	}

	if err := (&ansiblePulp).getCurrentDBPVC(clientset); err != nil {
		return
	}

	if err := (&ansiblePulp).getCurrentDBService(clientset); err != nil {
		return
	}

	if err := (&ansiblePulp).getCurrentDBSts(clientset); err != nil {
		return
	}

	csvName, err := ansiblePulp.getCurrentCSV(clientset)
	if err != nil {
		return
	}

	if err := ansiblePulp.deleteSubscription(clientset); err != nil {
		return
	}

	if err := ansiblePulp.deleteCSV(clientset, csvName); err != nil {
		return
	}

	if err := ansiblePulp.deleteDeployments(clientset); err != nil {
		return
	}

	if err := ansiblePulp.downscaleDBReplicas(clientset); err != nil {
		return
	}

	if err := ansiblePulp.updateDBService(clientset); err != nil {
		return
	}

	if err := ansiblePulp.subscribe(clientset); err != nil {
		return
	}

	if err := ansiblePulp.convert(clientset); err != nil {
		return
	} else {
		fmt.Println("✅ Migration finished")
	}

}
