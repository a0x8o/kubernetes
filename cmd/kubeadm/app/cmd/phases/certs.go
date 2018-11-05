/*
Copyright 2017 The Kubernetes Authors.

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

package phases

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmapiv1beta1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta1"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/options"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	certsphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/certs"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	"k8s.io/kubernetes/pkg/util/normalizer"
)

var (
	saKeyLongDesc = fmt.Sprintf(normalizer.LongDesc(`
		Generates the private key for signing service account tokens along with its public key, and saves them into
		%s and %s files.
		If both files already exist, kubeadm skips the generation step and existing files will be used.
		`+cmdutil.AlphaDisclaimer), kubeadmconstants.ServiceAccountPrivateKeyName, kubeadmconstants.ServiceAccountPublicKeyName)

	genericLongDesc = normalizer.LongDesc(`
		Generates the %[1]s, and saves them into %[2]s.cert and %[2]s.key files.%[3]s

		If both files already exist, kubeadm skips the generation step and existing files will be used.
		` + cmdutil.AlphaDisclaimer)
)

// certsData defines the behavior that a runtime data struct passed to the certs phase should
// have. Please note that we are using an interface in order to make this phase reusable in different workflows
// (and thus with different runtime data struct, all of them requested to be compliant to this interface)
type certsData interface {
	Cfg() *kubeadmapi.InitConfiguration
	ExternalCA() bool
	CertificateDir() string
	CertificateWriteDir() string
}

// NewCertsPhase returns the phase for the certs
func NewCertsPhase() workflow.Phase {
	return workflow.Phase{
		Name:     "certs",
		Short:    "Certificate generation",
		Phases:   newCertSubPhases(),
		Run:      runCerts,
		CmdFlags: getCertPhaseFlags("all"),
	}
}

// newCertSubPhases returns sub phases for certs phase
func newCertSubPhases() []workflow.Phase {
	subPhases := []workflow.Phase{}

	certTree, _ := certsphase.GetDefaultCertList().AsMap().CertTree()

	for ca, certList := range certTree {
		caPhase := newCertSubPhase(ca, runCAPhase(ca))
		subPhases = append(subPhases, caPhase)

		for _, cert := range certList {
			certPhase := newCertSubPhase(cert, runCertPhase(cert, ca))
			subPhases = append(subPhases, certPhase)
		}
	}

	// SA creates the private/public key pair, which doesn't use x509 at all
	saPhase := workflow.Phase{
		Name:  "sa",
		Short: "Generates a private key for signing service account tokens along with its public key",
		Long:  saKeyLongDesc,
		Run:   runCertsSa,
	}

	subPhases = append(subPhases, saPhase)

	return subPhases
}

func newCertSubPhase(certSpec *certsphase.KubeadmCert, run func(c workflow.RunData) error) workflow.Phase {
	phase := workflow.Phase{
		Name:  certSpec.Name,
		Short: fmt.Sprintf("Generates the %s", certSpec.LongName),
		Long: fmt.Sprintf(
			genericLongDesc,
			certSpec.LongName,
			certSpec.BaseName,
			getSANDescription(certSpec),
		),
		Run:      run,
		CmdFlags: getCertPhaseFlags(certSpec.Name),
	}
	return phase
}

func getCertPhaseFlags(name string) []string {
	flags := []string{
		options.CertificatesDir,
		options.CfgPath,
	}
	if name == "all" || name == "apiserver" {
		flags = append(flags,
			options.APIServerAdvertiseAddress,
			options.APIServerCertSANs,
			options.NetworkingDNSDomain,
			options.NetworkingServiceSubnet,
		)
	}
	return flags
}

func getSANDescription(certSpec *certsphase.KubeadmCert) string {
	//Defaulted config we will use to get SAN certs
	defaultConfig := &kubeadmapiv1beta1.InitConfiguration{
		APIEndpoint: kubeadmapiv1beta1.APIEndpoint{
			// GetAPIServerAltNames errors without an AdvertiseAddress; this is as good as any.
			AdvertiseAddress: "127.0.0.1",
		},
	}
	defaultInternalConfig := &kubeadmapi.InitConfiguration{}

	kubeadmscheme.Scheme.Default(defaultConfig)
	kubeadmscheme.Scheme.Convert(defaultConfig, defaultInternalConfig, nil)

	certConfig, err := certSpec.GetConfig(defaultInternalConfig)
	kubeadmutil.CheckErr(err)

	if len(certConfig.AltNames.DNSNames) == 0 && len(certConfig.AltNames.IPs) == 0 {
		return ""
	}
	// This mutates the certConfig, but we're throwing it after we construct the command anyway
	sans := []string{}

	for _, dnsName := range certConfig.AltNames.DNSNames {
		if dnsName != "" {
			sans = append(sans, dnsName)
		}
	}

	for _, ip := range certConfig.AltNames.IPs {
		sans = append(sans, ip.String())
	}
	return fmt.Sprintf("\n\nDefault SANs are %s", strings.Join(sans, ", "))
}

func runCertsSa(c workflow.RunData) error {
	data, ok := c.(certsData)
	if !ok {
		return errors.New("certs phase invoked with an invalid data struct")
	}

	// if external CA mode, skip service account key generation
	if data.ExternalCA() {
		fmt.Printf("[certs] External CA mode: Using existing sa keys\n")
		return nil
	}

	// if dryrunning, write certificates to a temporary folder (and defer restore to the path originally specified by the user)
	cfg := data.Cfg()
	cfg.CertificatesDir = data.CertificateWriteDir()
	defer func() { cfg.CertificatesDir = data.CertificateDir() }()

	// create the new service account key (or use existing)
	return certsphase.CreateServiceAccountKeyAndPublicKeyFiles(cfg)
}

func runCerts(c workflow.RunData) error {
	data, ok := c.(certsData)
	if !ok {
		return errors.New("certs phase invoked with an invalid data struct")
	}

	fmt.Printf("[certs] Using certificateDir folder %q\n", data.CertificateWriteDir())
	return nil
}

func runCAPhase(ca *certsphase.KubeadmCert) func(c workflow.RunData) error {
	return func(c workflow.RunData) error {
		data, ok := c.(certsData)
		if !ok {
			return errors.New("certs phase invoked with an invalid data struct")
		}

		// if external CA mode, skips certificate authority generation
		if data.ExternalCA() {
			fmt.Printf("[certs] External CA mode: Using existing %s certificate authority\n", ca.BaseName)
			return nil
		}

		// if using external etcd, skips etcd certificate authority generation
		if data.Cfg().Etcd.External != nil && ca.Name == "etcd-ca" {
			fmt.Printf("[certs] External etcd mode: Skipping %s certificate authority generation\n", ca.BaseName)
			return nil
		}

		// if dryrunning, write certificates authority to a temporary folder (and defer restore to the path originally specified by the user)
		cfg := data.Cfg()
		cfg.CertificatesDir = data.CertificateWriteDir()
		defer func() { cfg.CertificatesDir = data.CertificateDir() }()

		// create the new certificate authority (or use existing)
		return certsphase.CreateCACertAndKeyFiles(ca, cfg)
	}
}

func runCertPhase(cert *certsphase.KubeadmCert, caCert *certsphase.KubeadmCert) func(c workflow.RunData) error {
	return func(c workflow.RunData) error {
		data, ok := c.(certsData)
		if !ok {
			return errors.New("certs phase invoked with an invalid data struct")
		}

		// if external CA mode, skip certificate generation
		if data.ExternalCA() {
			fmt.Printf("[certs] External CA mode: Using existing %s certificate\n", cert.BaseName)
			return nil
		}

		// if using external etcd, skips etcd certificates generation
		if data.Cfg().Etcd.External != nil && cert.CAName == "etcd-ca" {
			fmt.Printf("[certs] External etcd mode: Skipping %s certificate authority generation\n", cert.BaseName)
			return nil
		}

		// if dryrunning, write certificates to a temporary folder (and defer restore to the path originally specified by the user)
		cfg := data.Cfg()
		cfg.CertificatesDir = data.CertificateWriteDir()
		defer func() { cfg.CertificatesDir = data.CertificateDir() }()

		// create the new certificate (or use existing)
		return certsphase.CreateCertAndKeyFilesWithCA(cert, caCert, cfg)
	}
}
