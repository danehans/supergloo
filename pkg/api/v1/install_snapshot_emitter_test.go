// Code generated by protoc-gen-solo-kit. DO NOT EDIT.

package v1

import (
	"context"
	"os"
	"path/filepath"
	"time"

	encryption_istio_io "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	kuberc "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/utils/log"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/setup"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = Describe("V1Emitter", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		namespace1               string
		namespace2               string
		cfg                      *rest.Config
		emitter                  InstallEmitter
		installClient            InstallClient
		istioCacertsSecretClient encryption_istio_io.IstioCacertsSecretClient
	)

	BeforeEach(func() {
		namespace1 = helpers.RandString(8)
		namespace2 = helpers.RandString(8)
		err := setup.SetupKubeForTest(namespace1)
		Expect(err).NotTo(HaveOccurred())
		err = setup.SetupKubeForTest(namespace2)
		kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		Expect(err).NotTo(HaveOccurred())

		cache := kuberc.NewKubeCache()
		// Install Constructor
		installClientFactory := &factory.KubeResourceClientFactory{
			Crd:         InstallCrd,
			Cfg:         cfg,
			SharedCache: cache,
		}
		installClient, err = NewInstallClient(installClientFactory)
		Expect(err).NotTo(HaveOccurred())

		// IstioCacertsSecret Constructor

		kube, err = kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		istioCacertsSecretClientFactory := &factory.KubeConfigMapClientFactory{
			Clientset: kube,
		}
		istioCacertsSecretClient, err = encryption_istio_io.NewIstioCacertsSecretClient(istioCacertsSecretClientFactory)
		Expect(err).NotTo(HaveOccurred())
		emitter = NewInstallEmitter(installClient, istioCacertsSecretClient)
	})
	AfterEach(func() {
		setup.TeardownKube(namespace1)
		setup.TeardownKube(namespace2)
	})
	It("tracks snapshots on changes to any resource", func() {
		ctx := context.Background()
		err := emitter.Register()
		Expect(err).NotTo(HaveOccurred())

		snapshots, errs, err := emitter.Snapshots([]string{namespace1, namespace2}, clients.WatchOpts{
			Ctx:         ctx,
			RefreshRate: time.Second,
		})
		Expect(err).NotTo(HaveOccurred())

		var snap *InstallSnapshot

		/*
			Install
		*/

		assertSnapshotInstalls := func(expectInstalls InstallList, unexpectInstalls InstallList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectInstalls {
						if _, err := snap.Installs.List().Find(expected.Metadata.Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectInstalls {
						if _, err := snap.Installs.List().Find(unexpected.Metadata.Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := installClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := installClient.List(namespace2, clients.ListOpts{})
					combined := nsList1.ByNamespace()
					combined.Add(nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}

		install1a, err := installClient.Write(NewInstall(namespace1, "angela"), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		install1b, err := installClient.Write(NewInstall(namespace2, "angela"), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotInstalls(InstallList{install1a, install1b}, nil)

		install2a, err := installClient.Write(NewInstall(namespace1, "bob"), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		install2b, err := installClient.Write(NewInstall(namespace2, "bob"), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotInstalls(InstallList{install1a, install1b, install2a, install2b}, nil)

		err = installClient.Delete(install2a.Metadata.Namespace, install2a.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = installClient.Delete(install2b.Metadata.Namespace, install2b.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotInstalls(InstallList{install1a, install1b}, InstallList{install2a, install2b})

		err = installClient.Delete(install1a.Metadata.Namespace, install1a.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = installClient.Delete(install1b.Metadata.Namespace, install1b.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotInstalls(nil, InstallList{install1a, install1b, install2a, install2b})

		/*
			IstioCacertsSecret
		*/

		assertSnapshotIstiocerts := func(expectIstiocerts encryption_istio_io.IstioCacertsSecretList, unexpectIstiocerts encryption_istio_io.IstioCacertsSecretList) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expectIstiocerts {
						if _, err := snap.Istiocerts.List().Find(expected.Metadata.Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpectIstiocerts {
						if _, err := snap.Istiocerts.List().Find(unexpected.Metadata.Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := istioCacertsSecretClient.List(namespace1, clients.ListOpts{})
					nsList2, _ := istioCacertsSecretClient.List(namespace2, clients.ListOpts{})
					combined := nsList1.ByNamespace()
					combined.Add(nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}

		istioCacertsSecret1a, err := istioCacertsSecretClient.Write(encryption_istio_io.NewIstioCacertsSecret(namespace1, "angela"), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		istioCacertsSecret1b, err := istioCacertsSecretClient.Write(encryption_istio_io.NewIstioCacertsSecret(namespace2, "angela"), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotIstiocerts(encryption_istio_io.IstioCacertsSecretList{istioCacertsSecret1a, istioCacertsSecret1b}, nil)

		istioCacertsSecret2a, err := istioCacertsSecretClient.Write(encryption_istio_io.NewIstioCacertsSecret(namespace1, "bob"), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		istioCacertsSecret2b, err := istioCacertsSecretClient.Write(encryption_istio_io.NewIstioCacertsSecret(namespace2, "bob"), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotIstiocerts(encryption_istio_io.IstioCacertsSecretList{istioCacertsSecret1a, istioCacertsSecret1b, istioCacertsSecret2a, istioCacertsSecret2b}, nil)

		err = istioCacertsSecretClient.Delete(istioCacertsSecret2a.Metadata.Namespace, istioCacertsSecret2a.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = istioCacertsSecretClient.Delete(istioCacertsSecret2b.Metadata.Namespace, istioCacertsSecret2b.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotIstiocerts(encryption_istio_io.IstioCacertsSecretList{istioCacertsSecret1a, istioCacertsSecret1b}, encryption_istio_io.IstioCacertsSecretList{istioCacertsSecret2a, istioCacertsSecret2b})

		err = istioCacertsSecretClient.Delete(istioCacertsSecret1a.Metadata.Namespace, istioCacertsSecret1a.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = istioCacertsSecretClient.Delete(istioCacertsSecret1b.Metadata.Namespace, istioCacertsSecret1b.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshotIstiocerts(nil, encryption_istio_io.IstioCacertsSecretList{istioCacertsSecret1a, istioCacertsSecret1b, istioCacertsSecret2a, istioCacertsSecret2b})
	})
})
