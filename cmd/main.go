package main

import (
	"context"
	"crypto/tls"
	"flag"
	"os"

	// Register auth plugins for out-of-cluster kubeconfigs.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	configv1 "github.com/openshift/api/config/v1"
	consolev1 "github.com/openshift/api/console/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	routev1 "github.com/openshift/api/route/v1"

	consoleuiv1alpha1 "github.com/kubevirt-ui/kubevirt-ui-operator/api/v1alpha1"
	"github.com/kubevirt-ui/kubevirt-ui-operator/controllers"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/hcoversion"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/operconfig"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/platform"
	"github.com/kubevirt-ui/kubevirt-ui-operator/internal/tlsprofile"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	utilruntime.Must(consoleuiv1alpha1.AddToScheme(scheme))
	utilruntime.Must(hcoversion.AddToScheme(scheme))
	utilruntime.Must(configv1.AddToScheme(scheme))
	utilruntime.Must(consolev1.AddToScheme(scheme))
	utilruntime.Must(operatorv1.AddToScheme(scheme))
	utilruntime.Must(routev1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var probeAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8443", "The address the metrics endpoint binds to (HTTPS). Use 0 to disable.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")
	opts := zap.Options{
		Development: false,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	opCfg := operconfig.LoadFromEnv()

	restConfig := ctrl.GetConfigOrDie()
	// Platform and HyperConverged detection must run before ctrl.NewManager: operconfig.CacheOptions
	// adds the HyperConverged informer conditionally, and controller-runtime resolves ByObject REST
	// mappings eagerly at manager construction, which fails if the CRD is not installed.
	startupClient, err := client.New(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		setupLog.Error(err, "unable to create startup client")
		os.Exit(1)
	}
	if ok, err := platform.IsOpenShift(context.Background(), startupClient); err != nil || !ok {
		setupLog.Error(err, "OpenShift console APIs not found; this operator requires OpenShift")
		os.Exit(1)
	}
	hcoInstalled, err := hcoversion.Installed(context.Background(), startupClient)
	if err != nil {
		setupLog.Error(err, "unable to detect HyperConverged CRD")
		os.Exit(1)
	}
	opCfg.HCOInstalled = hcoInstalled

	setupLog.Info("operator configuration",
		"operandNamespace", opCfg.OperandNamespace,
		"hcoNamespace", opCfg.HCONamespace,
		"hcoInstalled", opCfg.HCOInstalled,
		"pluginImage", opCfg.PluginImage,
		"proxyImage", opCfg.ProxyImage,
	)

	// Seed the metrics TLS settings from the cluster profile so the first handshake already honours it; the
	// reconciler keeps the holder current afterwards.
	metricsTLS := tlsprofile.NewHolder(tlsprofile.Resolve(nil))
	if s, err := tlsprofile.FromAPIServer(context.Background(), startupClient); err != nil {
		setupLog.Error(err, "unable to read APIServer TLS profile; metrics endpoint uses the Intermediate profile until the first reconcile")
	} else {
		metricsTLS.Store(s)
	}

	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme: scheme,
		Cache:  operconfig.CacheOptions(opCfg),
		Metrics: metricsserver.Options{
			BindAddress:    metricsAddr,
			SecureServing:  true,
			FilterProvider: filters.WithAuthenticationAndAuthorization,
			TLSOpts:        []func(*tls.Config){metricsTLS.TLSOpt()},
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "consoleui.kubevirt.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.KubeVirtConsoleUIReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		Config:     opCfg,
		MetricsTLS: metricsTLS,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KubeVirtConsoleUI")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
