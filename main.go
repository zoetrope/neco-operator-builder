package main

import (
	"fmt"
	"runtime"

	golangv3 "github.com/cybozu-go/neco-operator-builder/pkg/plugins/golang/v3"
	"github.com/cybozu-go/neco-operator-builder/pkg/version"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/kubebuilder/v3/pkg/cli"
	cfgv3 "sigs.k8s.io/kubebuilder/v3/pkg/config/v3"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	kustomizecommonv1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/common/kustomize/v1"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang"
	declarativev1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/declarative/v1"
)

func main() {
	// Bundle plugin which built the golang projects scaffold by Kubebuilder go/v3
	gov3Bundle, _ := plugin.NewBundle(golang.DefaultNameQualifier, plugin.Version{Number: 3},
		kustomizecommonv1.Plugin{},
		golangv3.Plugin{},
	)

	c, err := cli.New(
		cli.WithCommandName("neco-operator-builder"),
		cli.WithVersion(versionString()),
		cli.WithPlugins(
			gov3Bundle,
			&kustomizecommonv1.Plugin{},
			&declarativev1.Plugin{},
		),
		cli.WithDefaultPlugins(cfgv3.Version, gov3Bundle),
		cli.WithDefaultProjectVersion(cfgv3.Version),
		cli.WithCompletion(),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
}

// versionString returns the CLI version
func versionString() string {
	return fmt.Sprintf("neco-operator-builder version: %q, commit: %q, go version: %q, GOOS: %q, GOARCH: %q\n",
		version.GitVersion, version.GitCommit, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
