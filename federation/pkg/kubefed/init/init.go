/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// TODO(madhusdancs):
// 1. Make printSuccess prepend protocol/scheme to the IPs/hostnames.
// 1. Add a dry-run support.
// 2. Make all the API object names customizable.
//    Ex: federation-apiserver, federation-controller-manager, etc.
// 3. Make image name and tag customizable.
// 4. Separate etcd container from API server pod as a first step towards enabling HA.
// 5. Generate credentials of the following types for the API server:
//    i.  "known_tokens.csv"
//    ii. "basic_auth.csv"
// 6. Add the ability to customize DNS domain suffix. It should probably be derived
//    from cluster config.
// 7. Make etcd PVC size configurable.
// 8. Make API server and controller manager replicas customizable via the HA work.
package init

import (
	"fmt"
	"io"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	certutil "k8s.io/client-go/util/cert"
	triple "k8s.io/client-go/util/cert/triple"
	kubeadmkubeconfigphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/kubeconfig"
	"k8s.io/kubernetes/federation/pkg/kubefed/util"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/apis/rbac"
	client "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/version"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	APIServerCN                 = "federation-apiserver"
	ControllerManagerCN         = "federation-controller-manager"
	AdminCN                     = "admin"
	HostClusterLocalDNSZoneName = "cluster.local."

	// User name used by federation controller manager to make
	// calls to federation API server.
	ControllerManagerUser = "federation-controller-manager"

	// Name of the ServiceAccount used by the federation controller manager
	// to access the secrets in the host cluster.
	ControllerManagerSA = "federation-controller-manager"

	// Group name of the legacy/core API group
	legacyAPIGroup = ""

	lbAddrRetryInterval = 5 * time.Second
	podWaitInterval     = 2 * time.Second

	apiserverServiceTypeFlag      = "api-server-service-type"
	apiserverAdvertiseAddressFlag = "api-server-advertise-address"
)

var (
	init_long = templates.LongDesc(`
		Initialize a federation control plane.

        Federation control plane is hosted inside a Kubernetes
        cluster. The host cluster must be specified using the
        --host-cluster-context flag.`)
	init_example = templates.Examples(`
		# Initialize federation control plane for a federation
		# named foo in the host cluster whose local kubeconfig
		# context is bar.
		kubectl init foo --host-cluster-context=bar`)

	componentLabel = map[string]string{
		"app": "federated-cluster",
	}

	apiserverSvcSelector = map[string]string{
		"app":    "federated-cluster",
		"module": "federation-apiserver",
	}

	apiserverPodLabels = map[string]string{
		"app":    "federated-cluster",
		"module": "federation-apiserver",
	}

	controllerManagerPodLabels = map[string]string{
		"app":    "federated-cluster",
		"module": "federation-controller-manager",
	}

	hyperkubeImageName = "gcr.io/google_containers/hyperkube-amd64"
)

type initFederation struct {
	commonOptions util.SubcommandOptions
	options       initFederationOptions
}

type initFederationOptions struct {
	dnsZoneName                      string
	image                            string
	dnsProvider                      string
	etcdPVCapacity                   string
	etcdPersistentStorage            bool
	dryRun                           bool
	storageBackend                   string
	apiServerOverridesString         string
	apiServerOverrides               map[string]string
	controllerManagerOverridesString string
	controllerManagerOverrides       map[string]string
	apiServerServiceTypeString       string
	apiServerServiceType             v1.ServiceType
	apiServerAdvertiseAddress        string
}

func (o *initFederationOptions) Bind(flags *pflag.FlagSet) {
	defaultImage := fmt.Sprintf("%s:%s", hyperkubeImageName, version.Get())

	flags.StringVar(&o.dnsZoneName, "dns-zone-name", "", "DNS suffix for this federation. Federated Service DNS names are published with this suffix.")
	flags.StringVar(&o.image, "image", defaultImage, "Image to use for federation API server and controller manager binaries.")
	flags.StringVar(&o.dnsProvider, "dns-provider", "google-clouddns", "Dns provider to be used for this deployment.")
	flags.StringVar(&o.etcdPVCapacity, "etcd-pv-capacity", "10Gi", "Size of persistent volume claim to be used for etcd.")
	flags.BoolVar(&o.etcdPersistentStorage, "etcd-persistent-storage", true, "Use persistent volume for etcd. Defaults to 'true'.")
	flags.BoolVar(&o.dryRun, "dry-run", false, "dry run without sending commands to server.")
	flags.StringVar(&o.storageBackend, "storage-backend", "etcd2", "The storage backend for persistence. Options: 'etcd2' (default), 'etcd3'.")
	flags.StringVar(&o.apiServerOverridesString, "apiserver-arg-overrides", "", "comma separated list of federation-apiserver arguments to override: Example \"--arg1=value1,--arg2=value2...\"")
	flags.StringVar(&o.controllerManagerOverridesString, "controllermanager-arg-overrides", "", "comma separated list of federation-controller-manager arguments to override: Example \"--arg1=value1,--arg2=value2...\"")
	flags.StringVar(&o.apiServerServiceTypeString, apiserverServiceTypeFlag, string(v1.ServiceTypeLoadBalancer), "The type of service to create for federation API server. Options: 'LoadBalancer' (default), 'NodePort'.")
	flags.StringVar(&o.apiServerAdvertiseAddress, apiserverAdvertiseAddressFlag, "", "Preferred address to advertise api server nodeport service. Valid only if '"+apiserverServiceTypeFlag+"=NodePort'.")
}

// NewCmdInit defines the `init` command that bootstraps a federation
// control plane inside a set of host clusters.
func NewCmdInit(cmdOut io.Writer, config util.AdminConfig) *cobra.Command {
	opts := &initFederation{}

	cmd := &cobra.Command{
		Use:     "init FEDERATION_NAME --host-cluster-context=HOST_CONTEXT",
		Short:   "init initializes a federation control plane",
		Long:    init_long,
		Example: init_example,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(opts.Complete(cmd, args))
			cmdutil.CheckErr(opts.Run(cmdOut, config))
		},
	}

	flags := cmd.Flags()
	opts.commonOptions.Bind(flags)
	opts.options.Bind(flags)

	return cmd
}

type entityKeyPairs struct {
	ca                *triple.KeyPair
	server            *triple.KeyPair
	controllerManager *triple.KeyPair
	admin             *triple.KeyPair
}

// Complete ensures that options are valid and marshals them if necessary.
func (i *initFederation) Complete(cmd *cobra.Command, args []string) error {
	err := i.commonOptions.SetName(cmd, args)
	if err != nil {
		return err
	}

	i.options.apiServerServiceType = v1.ServiceType(i.options.apiServerServiceTypeString)
	if i.options.apiServerServiceType != v1.ServiceTypeLoadBalancer && i.options.apiServerServiceType != v1.ServiceTypeNodePort {
		return fmt.Errorf("invalid %s: %s, should be either %s or %s", apiserverServiceTypeFlag, i.options.apiServerServiceType, v1.ServiceTypeLoadBalancer, v1.ServiceTypeNodePort)
	}
	if i.options.apiServerAdvertiseAddress != "" {
		ip := net.ParseIP(i.options.apiServerAdvertiseAddress)
		if ip == nil {
			return fmt.Errorf("invalid %s: %s, should be a valid ip address", apiserverAdvertiseAddressFlag, i.options.apiServerAdvertiseAddress)
		}
		if i.options.apiServerServiceType != v1.ServiceTypeNodePort {
			return fmt.Errorf("%s should be passed only with '%s=NodePort'", apiserverAdvertiseAddressFlag, apiserverServiceTypeFlag)
		}
	}

	i.options.apiServerOverrides, err = marshallOverrides(i.options.apiServerOverridesString)
	if err != nil {
		return fmt.Errorf("Error marshalling --apiserver-arg-overrides: %v", err)
	}
	i.options.controllerManagerOverrides, err = marshallOverrides(i.options.controllerManagerOverridesString)
	if err != nil {
		return fmt.Errorf("Error marshalling --controllermanager-arg-overrides: %v", err)
	}

	return nil
}

// Run initializes a federation control plane.
// See the design doc in https://github.com/kubernetes/kubernetes/pull/34484
// for details.
func (i *initFederation) Run(cmdOut io.Writer, config util.AdminConfig) error {
	hostFactory := config.HostFactory(i.commonOptions.Host, i.commonOptions.Kubeconfig)
	hostClientset, err := hostFactory.ClientSet()
	if err != nil {
		return err
	}

	serverName := fmt.Sprintf("%s-apiserver", i.commonOptions.Name)
	serverCredName := fmt.Sprintf("%s-credentials", serverName)
	cmName := fmt.Sprintf("%s-controller-manager", i.commonOptions.Name)
	cmKubeconfigName := fmt.Sprintf("%s-kubeconfig", cmName)

	// 1. Create a namespace for federation system components
	_, err = createNamespace(hostClientset, i.commonOptions.FederationSystemNamespace, i.options.dryRun)
	if err != nil {
		return err
	}

	// 2. Expose a network endpoint for the federation API server
	svc, ips, hostnames, err := createService(hostClientset, i.commonOptions.FederationSystemNamespace, serverName, i.options.apiServerAdvertiseAddress, i.options.apiServerServiceType, i.options.dryRun)
	if err != nil {
		return err
	}

	// 3. Generate TLS certificates and credentials
	entKeyPairs, err := genCerts(i.commonOptions.FederationSystemNamespace, i.commonOptions.Name, svc.Name, HostClusterLocalDNSZoneName, ips, hostnames)
	if err != nil {
		return err
	}

	_, err = createAPIServerCredentialsSecret(hostClientset, i.commonOptions.FederationSystemNamespace, serverCredName, entKeyPairs, i.options.dryRun)
	if err != nil {
		return err
	}

	// 4. Create a kubeconfig secret
	_, err = createControllerManagerKubeconfigSecret(hostClientset, i.commonOptions.FederationSystemNamespace, i.commonOptions.Name, svc.Name, cmKubeconfigName, entKeyPairs, i.options.dryRun)
	if err != nil {
		return err
	}

	// 5. Create a persistent volume and a claim to store the federation
	// API server's state. This is where federation API server's etcd
	// stores its data.
	var pvc *api.PersistentVolumeClaim
	if i.options.etcdPersistentStorage {
		pvc, err = createPVC(hostClientset, i.commonOptions.FederationSystemNamespace, svc.Name, i.options.etcdPVCapacity, i.options.dryRun)
		if err != nil {
			return err
		}
	}

	// Since only one IP address can be specified as advertise address,
	// we arbitrarily pick the first available IP address
	// Pick user provided apiserverAdvertiseAddress over other available IP addresses.
	advertiseAddress := i.options.apiServerAdvertiseAddress
	if advertiseAddress == "" && len(ips) > 0 {
		advertiseAddress = ips[0]
	}

	// 6. Create federation API server
	_, err = createAPIServer(hostClientset, i.commonOptions.FederationSystemNamespace, serverName, i.options.image, serverCredName, advertiseAddress, i.options.storageBackend, i.options.apiServerOverrides, pvc, i.options.dryRun)
	if err != nil {
		return err
	}

	// 7. Create federation controller manager
	// 7a. Create a service account in the host cluster for federation
	// controller manager.
	sa, err := createControllerManagerSA(hostClientset, i.commonOptions.FederationSystemNamespace, i.options.dryRun)
	if err != nil {
		return err
	}

	// 7b. Create RBAC role and role binding for federation controller
	// manager service account.
	_, _, err = createRoleBindings(hostClientset, i.commonOptions.FederationSystemNamespace, sa.Name, i.options.dryRun)
	if err != nil {
		return err
	}

	// 7c. Create federation controller manager deployment.
	_, err = createControllerManager(hostClientset, i.commonOptions.FederationSystemNamespace, i.commonOptions.Name, svc.Name, cmName, i.options.image, cmKubeconfigName, i.options.dnsZoneName, i.options.dnsProvider, sa.Name, i.options.controllerManagerOverrides, i.options.dryRun)
	if err != nil {
		return err
	}

	// Pick the first ip/hostname to update the api server endpoint in kubeconfig and also to give information to user
	// In case of NodePort Service for api server, ips are node external ips.
	endpoint := ""
	if len(ips) > 0 {
		endpoint = ips[0]
	} else if len(hostnames) > 0 {
		endpoint = hostnames[0]
	}
	// If the service is nodeport, need to append the port to endpoint as it is non-standard port
	if i.options.apiServerServiceType == v1.ServiceTypeNodePort {
		endpoint = endpoint + ":" + strconv.Itoa(int(svc.Spec.Ports[0].NodePort))
	}

	// 8. Write the federation API server endpoint info, credentials
	// and context to kubeconfig
	err = updateKubeconfig(config, i.commonOptions.Name, endpoint, i.commonOptions.Kubeconfig, entKeyPairs, i.options.dryRun)
	if err != nil {
		return err
	}

	if !i.options.dryRun {
		fedPods := []string{serverName, cmName}
		err = waitForPods(hostClientset, fedPods, i.commonOptions.FederationSystemNamespace)
		if err != nil {
			return err
		}
		err = waitSrvHealthy(config, i.commonOptions.Name, i.commonOptions.Kubeconfig)
		if err != nil {
			return err
		}
		return printSuccess(cmdOut, ips, hostnames, svc)
	}
	_, err = fmt.Fprintf(cmdOut, "Federation control plane runs (dry run)\n")
	return err
}

func createNamespace(clientset *client.Clientset, namespace string, dryRun bool) (*api.Namespace, error) {
	ns := &api.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	if dryRun {
		return ns, nil
	}

	return clientset.Core().Namespaces().Create(ns)
}

func createService(clientset *client.Clientset, namespace, svcName, apiserverAdvertiseAddress string, apiserverServiceType v1.ServiceType, dryRun bool) (*api.Service, []string, []string, error) {
	svc := &api.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: namespace,
			Labels:    componentLabel,
		},
		Spec: api.ServiceSpec{
			Type:     api.ServiceType(apiserverServiceType),
			Selector: apiserverSvcSelector,
			Ports: []api.ServicePort{
				{
					Name:     "https",
					Protocol: "TCP",
					Port:     443,
				},
			},
		},
	}

	if dryRun {
		return svc, nil, nil, nil
	}

	var err error
	svc, err = clientset.Core().Services(namespace).Create(svc)

	ips := []string{}
	hostnames := []string{}
	if apiserverServiceType == v1.ServiceTypeLoadBalancer {
		ips, hostnames, err = waitForLoadBalancerAddress(clientset, svc, dryRun)
	} else {
		if apiserverAdvertiseAddress != "" {
			ips = append(ips, apiserverAdvertiseAddress)
		} else {
			ips, err = getClusterNodeIPs(clientset)
		}
	}
	if err != nil {
		return svc, nil, nil, err
	}

	return svc, ips, hostnames, err
}

func getClusterNodeIPs(clientset *client.Clientset) ([]string, error) {
	preferredAddressTypes := []api.NodeAddressType{
		api.NodeExternalIP,
		api.NodeLegacyHostIP,
	}
	nodeList, err := clientset.Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	nodeAddresses := []string{}
	for _, node := range nodeList.Items {
	OuterLoop:
		for _, addressType := range preferredAddressTypes {
			for _, address := range node.Status.Addresses {
				if address.Type == addressType {
					nodeAddresses = append(nodeAddresses, address.Address)
					break OuterLoop
				}
			}
		}
	}

	return nodeAddresses, nil
}

func waitForLoadBalancerAddress(clientset *client.Clientset, svc *api.Service, dryRun bool) ([]string, []string, error) {
	ips := []string{}
	hostnames := []string{}

	if dryRun {
		return ips, hostnames, nil
	}

	err := wait.PollImmediateInfinite(lbAddrRetryInterval, func() (bool, error) {
		pollSvc, err := clientset.Core().Services(svc.Namespace).Get(svc.Name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		if ings := pollSvc.Status.LoadBalancer.Ingress; len(ings) > 0 {
			for _, ing := range ings {
				if len(ing.IP) > 0 {
					ips = append(ips, ing.IP)
				}
				if len(ing.Hostname) > 0 {
					hostnames = append(hostnames, ing.Hostname)
				}
			}
			if len(ips) > 0 || len(hostnames) > 0 {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}

	return ips, hostnames, nil
}

func genCerts(svcNamespace, name, svcName, localDNSZoneName string, ips, hostnames []string) (*entityKeyPairs, error) {
	ca, err := triple.NewCA(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA key and certificate: %v", err)
	}
	server, err := triple.NewServerKeyPair(ca, APIServerCN, svcName, svcNamespace, localDNSZoneName, ips, hostnames)
	if err != nil {
		return nil, fmt.Errorf("failed to create federation API server key and certificate: %v", err)
	}
	cm, err := triple.NewClientKeyPair(ca, ControllerManagerCN, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create federation controller manager client key and certificate: %v", err)
	}
	admin, err := triple.NewClientKeyPair(ca, AdminCN, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client key and certificate for an admin: %v", err)
	}
	return &entityKeyPairs{
		ca:                ca,
		server:            server,
		controllerManager: cm,
		admin:             admin,
	}, nil
}

func createAPIServerCredentialsSecret(clientset *client.Clientset, namespace, credentialsName string, entKeyPairs *entityKeyPairs, dryRun bool) (*api.Secret, error) {
	// Build the secret object with API server credentials.
	secret := &api.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      credentialsName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"ca.crt":     certutil.EncodeCertPEM(entKeyPairs.ca.Cert),
			"server.crt": certutil.EncodeCertPEM(entKeyPairs.server.Cert),
			"server.key": certutil.EncodePrivateKeyPEM(entKeyPairs.server.Key),
		},
	}

	if dryRun {
		return secret, nil
	}
	// Boilerplate to create the secret in the host cluster.
	return clientset.Core().Secrets(namespace).Create(secret)
}

func createControllerManagerKubeconfigSecret(clientset *client.Clientset, namespace, name, svcName, kubeconfigName string, entKeyPairs *entityKeyPairs, dryRun bool) (*api.Secret, error) {
	config := kubeadmkubeconfigphase.MakeClientConfigWithCerts(
		fmt.Sprintf("https://%s", svcName),
		name,
		ControllerManagerUser,
		certutil.EncodeCertPEM(entKeyPairs.ca.Cert),
		certutil.EncodePrivateKeyPEM(entKeyPairs.controllerManager.Key),
		certutil.EncodeCertPEM(entKeyPairs.controllerManager.Cert),
	)

	return util.CreateKubeconfigSecret(clientset, config, namespace, kubeconfigName, dryRun)
}

func createPVC(clientset *client.Clientset, namespace, svcName, etcdPVCapacity string, dryRun bool) (*api.PersistentVolumeClaim, error) {
	capacity, err := resource.ParseQuantity(etcdPVCapacity)
	if err != nil {
		return nil, err
	}

	pvc := &api.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-etcd-claim", svcName),
			Namespace: namespace,
			Labels:    componentLabel,
			Annotations: map[string]string{
				"volume.alpha.kubernetes.io/storage-class": "yes",
			},
		},
		Spec: api.PersistentVolumeClaimSpec{
			AccessModes: []api.PersistentVolumeAccessMode{
				api.ReadWriteOnce,
			},
			Resources: api.ResourceRequirements{
				Requests: api.ResourceList{
					api.ResourceStorage: capacity,
				},
			},
		},
	}

	if dryRun {
		return pvc, nil
	}

	return clientset.Core().PersistentVolumeClaims(namespace).Create(pvc)
}

func createAPIServer(clientset *client.Clientset, namespace, name, image, credentialsName, advertiseAddress, storageBackend string, argOverrides map[string]string, pvc *api.PersistentVolumeClaim, dryRun bool) (*extensions.Deployment, error) {
	command := []string{
		"/hyperkube",
		"federation-apiserver",
	}
	argsMap := map[string]string{
		"--bind-address":         "0.0.0.0",
		"--etcd-servers":         "http://localhost:2379",
		"--secure-port":          "443",
		"--client-ca-file":       "/etc/federation/apiserver/ca.crt",
		"--tls-cert-file":        "/etc/federation/apiserver/server.crt",
		"--tls-private-key-file": "/etc/federation/apiserver/server.key",
		"--admission-control":    "NamespaceLifecycle",
	}

	argsMap["--storage-backend"] = storageBackend
	if advertiseAddress != "" {
		argsMap["--advertise-address"] = advertiseAddress
	}

	args := argMapsToArgStrings(argsMap, argOverrides)
	command = append(command, args...)

	dep := &extensions.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    componentLabel,
		},
		Spec: extensions.DeploymentSpec{
			Replicas: 1,
			Template: api.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   name,
					Labels: apiserverPodLabels,
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:    "apiserver",
							Image:   image,
							Command: command,
							Ports: []api.ContainerPort{
								{
									Name:          "https",
									ContainerPort: 443,
								},
								{
									Name:          "local",
									ContainerPort: 8080,
								},
							},
							VolumeMounts: []api.VolumeMount{
								{
									Name:      credentialsName,
									MountPath: "/etc/federation/apiserver",
									ReadOnly:  true,
								},
							},
						},
						{
							Name:  "etcd",
							Image: "gcr.io/google_containers/etcd:3.0.14-alpha.1",
							Command: []string{
								"/usr/local/bin/etcd",
								"--data-dir",
								"/var/etcd/data",
							},
						},
					},
					Volumes: []api.Volume{
						{
							Name: credentialsName,
							VolumeSource: api.VolumeSource{
								Secret: &api.SecretVolumeSource{
									SecretName: credentialsName,
								},
							},
						},
					},
				},
			},
		},
	}

	if pvc != nil {
		dataVolumeName := "etcddata"
		etcdVolume := api.Volume{
			Name: dataVolumeName,
			VolumeSource: api.VolumeSource{
				PersistentVolumeClaim: &api.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc.Name,
				},
			},
		}
		etcdVolumeMount := api.VolumeMount{
			Name:      dataVolumeName,
			MountPath: "/var/etcd",
		}

		dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes, etcdVolume)
		for i, container := range dep.Spec.Template.Spec.Containers {
			if container.Name == "etcd" {
				dep.Spec.Template.Spec.Containers[i].VolumeMounts = append(dep.Spec.Template.Spec.Containers[i].VolumeMounts, etcdVolumeMount)
			}
		}
	}

	if dryRun {
		return dep, nil
	}

	return clientset.Extensions().Deployments(namespace).Create(dep)
}

func createControllerManagerSA(clientset *client.Clientset, namespace string, dryRun bool) (*api.ServiceAccount, error) {
	sa := &api.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ControllerManagerSA,
			Namespace: namespace,
			Labels:    componentLabel,
		},
	}
	if dryRun {
		return sa, nil
	}
	return clientset.Core().ServiceAccounts(namespace).Create(sa)
}

func createRoleBindings(clientset *client.Clientset, namespace, saName string, dryRun bool) (*rbac.Role, *rbac.RoleBinding, error) {
	roleName := "federation-system:federation-controller-manager"
	role := &rbac.Role{
		// a role to use for bootstrapping the federation-controller-manager so it can access
		// secrets in the host cluster to access other clusters.
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: namespace,
			Labels:    componentLabel,
		},
		Rules: []rbac.PolicyRule{
			rbac.NewRule("get", "list", "watch").Groups(legacyAPIGroup).Resources("secrets").RuleOrDie(),
		},
	}

	rolebinding, err := rbac.NewRoleBinding(roleName, namespace).SAs(namespace, saName).Binding()
	if err != nil {
		return nil, nil, err
	}
	rolebinding.Labels = componentLabel

	if dryRun {
		return role, &rolebinding, nil
	}

	newRole, err := clientset.Rbac().Roles(namespace).Create(role)
	if err != nil {
		return nil, nil, err
	}

	newRolebinding, err := clientset.Rbac().RoleBindings(namespace).Create(&rolebinding)
	return newRole, newRolebinding, err
}

func createControllerManager(clientset *client.Clientset, namespace, name, svcName, cmName, image, kubeconfigName, dnsZoneName, dnsProvider, saName string, argOverrides map[string]string, dryRun bool) (*extensions.Deployment, error) {
	command := []string{
		"/hyperkube",
		"federation-controller-manager",
	}
	argsMap := map[string]string{
		"--kubeconfig":          "/etc/federation/controller-manager/kubeconfig",
		"--dns-provider-config": "",
	}

	argsMap["--master"] = fmt.Sprintf("https://%s", svcName)
	argsMap["--dns-provider"] = dnsProvider
	argsMap["--federation-name"] = name
	argsMap["--zone-name"] = dnsZoneName

	args := argMapsToArgStrings(argsMap, argOverrides)
	command = append(command, args...)

	dep := &extensions.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: namespace,
			Labels:    componentLabel,
		},
		Spec: extensions.DeploymentSpec{
			Replicas: 1,
			Template: api.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   cmName,
					Labels: controllerManagerPodLabels,
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:    "controller-manager",
							Image:   image,
							Command: command,
							VolumeMounts: []api.VolumeMount{
								{
									Name:      kubeconfigName,
									MountPath: "/etc/federation/controller-manager",
									ReadOnly:  true,
								},
							},
							Env: []api.EnvVar{
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &api.EnvVarSource{
										FieldRef: &api.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					Volumes: []api.Volume{
						{
							Name: kubeconfigName,
							VolumeSource: api.VolumeSource{
								Secret: &api.SecretVolumeSource{
									SecretName: kubeconfigName,
								},
							},
						},
					},
					ServiceAccountName: saName,
				},
			},
		},
	}

	if dryRun {
		return dep, nil
	}
	return clientset.Extensions().Deployments(namespace).Create(dep)
}

func marshallOverrides(overrideArgString string) (map[string]string, error) {
	if overrideArgString == "" {
		return nil, nil
	}

	argsMap := make(map[string]string)
	overrideArgs := strings.Split(overrideArgString, ",")
	for _, overrideArg := range overrideArgs {
		splitArg := strings.Split(overrideArg, "=")
		if len(splitArg) != 2 {
			return nil, fmt.Errorf("wrong format for override arg: %s", overrideArg)
		}
		key := strings.TrimSpace(splitArg[0])
		val := strings.TrimSpace(splitArg[1])
		if len(key) == 0 {
			return nil, fmt.Errorf("wrong format for override arg: %s, arg name cannot be empty", overrideArg)
		}
		argsMap[key] = val
	}
	return argsMap, nil
}

func argMapsToArgStrings(argsMap, overrides map[string]string) []string {
	for key, val := range overrides {
		argsMap[key] = val
	}
	args := []string{}
	for key, value := range argsMap {
		args = append(args, fmt.Sprintf("%s=%s", key, value))
	}
	// This is needed for the unit test deep copy to get an exact match
	sort.Strings(args)
	return args
}

func waitForPods(clientset *client.Clientset, fedPods []string, namespace string) error {
	err := wait.PollInfinite(podWaitInterval, func() (bool, error) {
		podCheck := len(fedPods)
		podList, err := clientset.Core().Pods(namespace).List(metav1.ListOptions{})
		if err != nil {
			return false, nil
		}
		for _, pod := range podList.Items {
			for _, fedPod := range fedPods {
				if strings.HasPrefix(pod.Name, fedPod) && pod.Status.Phase == "Running" {
					podCheck -= 1
				}
			}
			//ensure that all pods are in running state or keep waiting
			if podCheck == 0 {
				return true, nil
			}
		}
		return false, nil
	})
	return err
}

func waitSrvHealthy(config util.AdminConfig, context, kubeconfig string) error {
	fedClientSet, err := config.FederationClientset(context, kubeconfig)
	if err != nil {
		return err
	}
	fedDiscoveryClient := fedClientSet.Discovery()
	err = wait.PollInfinite(podWaitInterval, func() (bool, error) {
		body, err := fedDiscoveryClient.RESTClient().Get().AbsPath("/healthz").Do().Raw()
		if err != nil {
			return false, nil
		}
		if strings.EqualFold(string(body), "ok") {
			return true, nil
		}
		return false, nil
	})
	return err
}

func printSuccess(cmdOut io.Writer, ips, hostnames []string, svc *api.Service) error {
	svcEndpoints := append(ips, hostnames...)
	endpoints := strings.Join(svcEndpoints, ", ")
	if svc.Spec.Type == api.ServiceTypeNodePort {
		endpoints = ips[0] + ":" + strconv.Itoa(int(svc.Spec.Ports[0].NodePort))
		if len(ips) > 1 {
			endpoints = endpoints + ", ..."
		}
	}

	_, err := fmt.Fprintf(cmdOut, "Federation API server is running at: %s\n", endpoints)
	return err
}

func updateKubeconfig(config util.AdminConfig, name, endpoint, kubeConfigPath string, entKeyPairs *entityKeyPairs, dryRun bool) error {
	po := config.PathOptions()
	po.LoadingRules.ExplicitPath = kubeConfigPath
	kubeconfig, err := po.GetStartingConfig()
	if err != nil {
		return err
	}

	// Populate API server endpoint info.
	cluster := clientcmdapi.NewCluster()
	// Prefix "https" as the URL scheme to endpoint.
	if !strings.HasPrefix(endpoint, "https://") {
		endpoint = fmt.Sprintf("https://%s", endpoint)
	}
	cluster.Server = endpoint
	cluster.CertificateAuthorityData = certutil.EncodeCertPEM(entKeyPairs.ca.Cert)

	// Populate credentials.
	authInfo := clientcmdapi.NewAuthInfo()
	authInfo.ClientCertificateData = certutil.EncodeCertPEM(entKeyPairs.admin.Cert)
	authInfo.ClientKeyData = certutil.EncodePrivateKeyPEM(entKeyPairs.admin.Key)
	authInfo.Username = AdminCN

	// Populate context.
	context := clientcmdapi.NewContext()
	context.Cluster = name
	context.AuthInfo = name

	// Update the config struct with API server endpoint info,
	// credentials and context.
	kubeconfig.Clusters[name] = cluster
	kubeconfig.AuthInfos[name] = authInfo
	kubeconfig.Contexts[name] = context

	if !dryRun {
		// Write the update kubeconfig.
		if err := clientcmd.ModifyConfig(po, *kubeconfig, true); err != nil {
			return err
		}
	}

	return nil
}
