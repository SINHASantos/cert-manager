/*
Copyright 2020 The cert-manager Authors.

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

package certificaterequest

import (
	"context"
	"crypto/x509"
	"net"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/cert-manager/cert-manager/e2e-tests/framework"
	"github.com/cert-manager/cert-manager/e2e-tests/framework/addon"
	vaultaddon "github.com/cert-manager/cert-manager/e2e-tests/framework/addon/vault"
	"github.com/cert-manager/cert-manager/e2e-tests/framework/helper/validation/certificaterequests"
	"github.com/cert-manager/cert-manager/e2e-tests/util"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/cert-manager/cert-manager/test/unit/gen"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = framework.CertManagerDescribe("Vault Issuer CertificateRequest (AppRole)", func() {
	runVaultAppRoleTests(cmapi.IssuerKind)
})

var _ = framework.CertManagerDescribe("Vault ClusterIssuer CertificateRequest (AppRole)", func() {
	runVaultAppRoleTests(cmapi.ClusterIssuerKind)
})

func runVaultAppRoleTests(issuerKind string) {
	f := framework.NewDefaultFramework("create-vault-certificaterequest")
	h := f.Helper()

	var (
		crDNSNames    = []string{"dnsName1.co", "dnsName2.ninja"}
		crIPAddresses = []net.IP{
			[]byte{8, 8, 8, 8},
			[]byte{1, 1, 1, 1},
		}
	)

	certificateRequestName := "test-vault-certificaterequest"
	var vaultIssuerName string

	appRoleSecretGeneratorName := "vault-approle-secret-"
	var roleId, secretId string
	var vaultSecretName, vaultSecretNamespace string

	var setup *vaultaddon.VaultInitializer

	BeforeEach(func(testingCtx context.Context) {
		By("Configuring the Vault server")
		if issuerKind == cmapi.IssuerKind {
			vaultSecretNamespace = f.Namespace.Name
		} else {
			vaultSecretNamespace = f.Config.Addons.CertManager.ClusterResourceNamespace
		}

		setup = vaultaddon.NewVaultInitializerAppRole(
			addon.Base.Details().KubeClient,
			*addon.Vault.Details(),
			false,
		)
		Expect(setup.Init(testingCtx)).NotTo(HaveOccurred(), "failed to init vault")
		Expect(setup.Setup(testingCtx)).NotTo(HaveOccurred(), "failed to setup vault")

		var err error
		roleId, secretId, err = setup.CreateAppRole(testingCtx)
		Expect(err).NotTo(HaveOccurred())

		sec, err := f.KubeClientSet.CoreV1().Secrets(vaultSecretNamespace).Create(testingCtx, vaultaddon.NewVaultAppRoleSecret(appRoleSecretGeneratorName, secretId), metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
		vaultSecretName = sec.Name
	})

	JustAfterEach(func(testingCtx context.Context) {
		By("Cleaning up")
		Expect(setup.Clean(testingCtx)).NotTo(HaveOccurred())

		if issuerKind == cmapi.IssuerKind {
			err := f.CertManagerClientSet.CertmanagerV1().Issuers(f.Namespace.Name).Delete(testingCtx, vaultIssuerName, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		} else {
			err := f.CertManagerClientSet.CertmanagerV1().ClusterIssuers().Delete(testingCtx, vaultIssuerName, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		}

		err := f.KubeClientSet.CoreV1().Secrets(vaultSecretNamespace).Delete(testingCtx, vaultSecretName, metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should generate a new valid certificate", func(testingCtx context.Context) {
		By("Creating an Issuer")
		vaultURL := addon.Vault.Details().URL

		crClient := f.CertManagerClientSet.CertmanagerV1().CertificateRequests(f.Namespace.Name)

		var err error
		if issuerKind == cmapi.IssuerKind {
			vaultIssuer := gen.IssuerWithRandomName("test-vault-issuer-",
				gen.SetIssuerNamespace(f.Namespace.Name),
				gen.SetIssuerVaultURL(vaultURL),
				gen.SetIssuerVaultPath(setup.IntermediateSignPath()),
				gen.SetIssuerVaultCABundle(addon.Vault.Details().VaultCA),
				gen.SetIssuerVaultAppRoleAuth("secretkey", vaultSecretName, roleId, setup.AppRoleAuthPath()))
			iss, err := f.CertManagerClientSet.CertmanagerV1().Issuers(f.Namespace.Name).Create(testingCtx, vaultIssuer, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			vaultIssuerName = iss.Name
		} else {
			vaultIssuer := gen.ClusterIssuerWithRandomName("test-vault-issuer-",
				gen.SetIssuerVaultURL(vaultURL),
				gen.SetIssuerVaultPath(setup.IntermediateSignPath()),
				gen.SetIssuerVaultCABundle(addon.Vault.Details().VaultCA),
				gen.SetIssuerVaultAppRoleAuth("secretkey", vaultSecretName, roleId, setup.AppRoleAuthPath()))
			iss, err := f.CertManagerClientSet.CertmanagerV1().ClusterIssuers().Create(testingCtx, vaultIssuer, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			vaultIssuerName = iss.Name
		}

		By("Waiting for Issuer to become Ready")
		if issuerKind == cmapi.IssuerKind {
			err = util.WaitForIssuerCondition(testingCtx, f.CertManagerClientSet.CertmanagerV1().Issuers(f.Namespace.Name),
				vaultIssuerName,
				cmapi.IssuerCondition{
					Type:   cmapi.IssuerConditionReady,
					Status: cmmeta.ConditionTrue,
				})
		} else {
			err = util.WaitForClusterIssuerCondition(testingCtx, f.CertManagerClientSet.CertmanagerV1().ClusterIssuers(),
				vaultIssuerName,
				cmapi.IssuerCondition{
					Type:   cmapi.IssuerConditionReady,
					Status: cmmeta.ConditionTrue,
				})
		}
		Expect(err).NotTo(HaveOccurred())

		By("Creating a CertificateRequest")
		csr, key, err := gen.CSR(x509.RSA, gen.SetCSRCommonName(crDNSNames[0]), gen.SetCSRDNSNames(crDNSNames...), gen.SetCSRIPAddresses(crIPAddresses...))
		Expect(err).NotTo(HaveOccurred())
		cr := gen.CertificateRequest(certificateRequestName,
			gen.SetCertificateRequestNamespace(f.Namespace.Name),
			gen.SetCertificateRequestIssuer(cmmeta.ObjectReference{Kind: issuerKind, Name: vaultIssuerName}),
			gen.SetCertificateRequestDuration(&metav1.Duration{Duration: time.Hour * 24 * 90}),
			gen.SetCertificateRequestCSR(csr),
		)
		_, err = crClient.Create(testingCtx, cr, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())

		By("Verifying the Certificate is valid")
		err = h.WaitCertificateRequestIssuedValid(testingCtx, f.Namespace.Name, certificateRequestName, time.Minute*5, key)
		Expect(err).NotTo(HaveOccurred())
	})

	cases := []struct {
		inputDuration    *metav1.Duration
		expectedDuration time.Duration
		label            string
		event            string
	}{
		{
			inputDuration:    &metav1.Duration{Duration: time.Hour * 24 * 35},
			expectedDuration: time.Hour * 24 * 35,
			label:            "valid for 35 days",
		},
		{
			inputDuration:    nil,
			expectedDuration: time.Hour * 24 * 90,
			label:            "valid for the default value (90 days)",
		},
		{
			inputDuration:    &metav1.Duration{Duration: time.Hour * 24 * 365},
			expectedDuration: time.Hour * 24 * 90,
			label:            "with Vault configured maximum TTL duration (90 days) when requested duration is greater than TTL",
		},
		{
			inputDuration:    &metav1.Duration{Duration: time.Hour * 24 * 240},
			expectedDuration: time.Hour * 24 * 90,
			label:            "with a warning event when renewBefore is bigger than the duration",
		},
	}

	for _, v := range cases {
		It("should generate a new certificate "+v.label, func(testingCtx context.Context) {
			By("Creating an Issuer")

			var err error
			if issuerKind == cmapi.IssuerKind {
				vaultIssuer := gen.IssuerWithRandomName("test-vault-issuer-",
					gen.SetIssuerNamespace(f.Namespace.Name),
					gen.SetIssuerVaultURL(addon.Vault.Details().URL),
					gen.SetIssuerVaultPath(setup.IntermediateSignPath()),
					gen.SetIssuerVaultCABundle(addon.Vault.Details().VaultCA),
					gen.SetIssuerVaultAppRoleAuth("secretkey", vaultSecretName, roleId, setup.AppRoleAuthPath()))
				iss, err := f.CertManagerClientSet.CertmanagerV1().Issuers(f.Namespace.Name).Create(testingCtx, vaultIssuer, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())

				vaultIssuerName = iss.Name
			} else {
				vaultIssuer := gen.ClusterIssuerWithRandomName("test-vault-issuer-",
					gen.SetIssuerVaultURL(addon.Vault.Details().URL),
					gen.SetIssuerVaultPath(setup.IntermediateSignPath()),
					gen.SetIssuerVaultCABundle(addon.Vault.Details().VaultCA),
					gen.SetIssuerVaultAppRoleAuth("secretkey", vaultSecretName, roleId, setup.AppRoleAuthPath()))
				iss, err := f.CertManagerClientSet.CertmanagerV1().ClusterIssuers().Create(testingCtx, vaultIssuer, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())

				vaultIssuerName = iss.Name
			}

			By("Waiting for Issuer to become Ready")
			if issuerKind == cmapi.IssuerKind {
				err = util.WaitForIssuerCondition(testingCtx, f.CertManagerClientSet.CertmanagerV1().Issuers(f.Namespace.Name),
					vaultIssuerName,
					cmapi.IssuerCondition{
						Type:   cmapi.IssuerConditionReady,
						Status: cmmeta.ConditionTrue,
					})
			} else {
				err = util.WaitForClusterIssuerCondition(testingCtx, f.CertManagerClientSet.CertmanagerV1().ClusterIssuers(),
					vaultIssuerName,
					cmapi.IssuerCondition{
						Type:   cmapi.IssuerConditionReady,
						Status: cmmeta.ConditionTrue,
					})
			}
			Expect(err).NotTo(HaveOccurred())

			By("Creating a CertificateRequest")
			crClient := f.CertManagerClientSet.CertmanagerV1().CertificateRequests(f.Namespace.Name)

			csr, key, err := gen.CSR(x509.RSA, gen.SetCSRCommonName(crDNSNames[0]), gen.SetCSRDNSNames(crDNSNames...), gen.SetCSRIPAddresses(crIPAddresses...))
			Expect(err).NotTo(HaveOccurred())
			cr := gen.CertificateRequest(certificateRequestName,
				gen.SetCertificateRequestNamespace(f.Namespace.Name),
				gen.SetCertificateRequestIssuer(cmmeta.ObjectReference{Kind: issuerKind, Name: vaultIssuerName}),
				gen.SetCertificateRequestDuration(v.inputDuration),
				gen.SetCertificateRequestCSR(csr),
			)
			_, err = crClient.Create(testingCtx, cr, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			err = h.WaitCertificateRequestIssuedValid(testingCtx, f.Namespace.Name, certificateRequestName, time.Minute*5, key)
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the Certificate is valid")
			_, err = crClient.Get(testingCtx, cr.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			// Vault can issue certificates with slightly skewed duration.
			err = h.ValidateCertificateRequest(types.NamespacedName{
				Namespace: f.Namespace.Name,
				Name:      certificateRequestName,
			}, key, certificaterequests.ExpectDuration(v.expectedDuration, 30*time.Second))
			Expect(err).NotTo(HaveOccurred())
		})
	}
}
