package cmd

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Run{}

// Run scaffolds a file that defines the controller manager entry point
type Run struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.DomainMixin
	machinery.RepositoryMixin
	machinery.ComponentConfigMixin
	machinery.ProjectNameMixin
}

// SetTemplateDefaults implements file.Template
func (f *Run) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("cmd", f.ProjectName+"-controller", "sub", "run.go")
	}

	f.TemplateBody = fmt.Sprintf(runTemplate,
		machinery.NewMarkerFor(f.Path, importMarker),
		machinery.NewMarkerFor(f.Path, addSchemeMarker),
		machinery.NewMarkerFor(f.Path, setupMarker),
	)

	return nil
}

var _ machinery.Inserter = &RunUpdater{}

func NewRunUpdater(projectName string, writeResource, writeController bool) *RunUpdater{
	return &RunUpdater{
		PathMixin: machinery.PathMixin{
			Path: filepath.Join("cmd", projectName+"-controller", "sub", "run.go"),
		},
		WireResource: writeResource,
		WireController: writeController,
	}
}

// RunUpdater updates run.go to run Controllers
type RunUpdater struct { //nolint:maligned
	machinery.PathMixin
	machinery.RepositoryMixin
	machinery.MultiGroupMixin
	machinery.ResourceMixin

	// Flags to indicate which parts need to be included when updating the file
	WireResource, WireController, WireWebhook bool
}

// GetPath implements file.Builder
func (f *RunUpdater) GetPath() string {
	return f.Path
}

// GetIfExistsAction implements file.Builder
func (*RunUpdater) GetIfExistsAction() machinery.IfExistsAction {
	return machinery.OverwriteFile
}

const (
	importMarker    = "imports"
	addSchemeMarker = "scheme"
	setupMarker     = "builder"
)

// GetMarkers implements file.Inserter
func (f *RunUpdater) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(f.Path, importMarker),
		machinery.NewMarkerFor(f.Path, addSchemeMarker),
		machinery.NewMarkerFor(f.Path, setupMarker),
	}
}

const (
	apiImportCodeFragment = `%s "%s"
`
	controllerImportCodeFragment = `"%s/controllers"
`
	multiGroupControllerImportCodeFragment = `%scontrollers "%s/controllers/%s"
`
	addschemeCodeFragment = `utilruntime.Must(%s.AddToScheme(scheme))
`
	reconcilerSetupCodeFragment = `if err = (&controllers.%sReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "%s")
		os.Exit(1)
	}
`
	multiGroupReconcilerSetupCodeFragment = `if err = (&%scontrollers.%sReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "%s")
		os.Exit(1)
	}
`
	webhookSetupCodeFragment = `if err = (&%s.%s{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "%s")
		os.Exit(1)
	}
`
)

// GetCodeFragments implements file.Inserter
func (f *RunUpdater) GetCodeFragments() machinery.CodeFragmentsMap {
	fragments := make(machinery.CodeFragmentsMap, 3)

	// If resource is not being provided we are creating the file, not updating it
	if f.Resource == nil {
		return fragments
	}

	// Generate import code fragments
	imports := make([]string, 0)
	if f.WireResource {
		imports = append(imports, fmt.Sprintf(apiImportCodeFragment, f.Resource.ImportAlias(), f.Resource.Path))
	}

	if f.WireController {
		if !f.MultiGroup || f.Resource.Group == "" {
			imports = append(imports, fmt.Sprintf(controllerImportCodeFragment, f.Repo))
		} else {
			imports = append(imports, fmt.Sprintf(multiGroupControllerImportCodeFragment,
				f.Resource.PackageName(), f.Repo, f.Resource.Group))
		}
	}

	// Generate add scheme code fragments
	addScheme := make([]string, 0)
	if f.WireResource {
		addScheme = append(addScheme, fmt.Sprintf(addschemeCodeFragment, f.Resource.ImportAlias()))
	}

	// Generate setup code fragments
	setup := make([]string, 0)
	if f.WireController {
		if !f.MultiGroup || f.Resource.Group == "" {
			setup = append(setup, fmt.Sprintf(reconcilerSetupCodeFragment,
				f.Resource.Kind, f.Resource.Kind))
		} else {
			setup = append(setup, fmt.Sprintf(multiGroupReconcilerSetupCodeFragment,
				f.Resource.PackageName(), f.Resource.Kind, f.Resource.Kind))
		}
	}
	if f.WireWebhook {
		setup = append(setup, fmt.Sprintf(webhookSetupCodeFragment,
			f.Resource.ImportAlias(), f.Resource.Kind, f.Resource.Kind))
	}

	// Only store code fragments in the map if the slices are non-empty
	if len(imports) != 0 {
		fragments[machinery.NewMarkerFor(f.Path, importMarker)] = imports
	}
	if len(addScheme) != 0 {
		fragments[machinery.NewMarkerFor(f.Path, addSchemeMarker)] = addScheme
	}
	if len(setup) != 0 {
		fragments[machinery.NewMarkerFor(f.Path, setupMarker)] = setup
	}

	return fragments
}

var runTemplate = `{{ .Boilerplate }}

package sub

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	%s
)

var (
	scheme = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	%s
}

func main() {
{{- if not .ComponentConfig }}
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. " +
		"Enabling this will ensure there is only one active controller manager.")
{{- else }}
  var configFile string
	flag.StringVar(&configFile, "config", "", 
		"The controller will load its initial configuration from this file. " +
		"Omit this flag to use the default configuration values. " +
		"Command-line flags override configuration from this file.")
{{- end }}
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

{{ if not .ComponentConfig }}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "{{ hashFNV .Repo }}.{{ .Domain }}",
	})
{{- else }}
	var err error
	options := ctrl.Options{Scheme: scheme}
	if configFile != "" {
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile))
		if err != nil {
			setupLog.Error(err, "unable to load the config file")
			os.Exit(1)
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
{{- end }}
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	%s

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
`
