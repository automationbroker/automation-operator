package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/automationbroker/automation-operator/pkg/crd"
	stub "github.com/automationbroker/automation-operator/pkg/handler"
	"github.com/automationbroker/bundle-lib/bundle"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	k8sutil "github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/sirupsen/logrus"
)

func init() {
	pflag.Int("resync", 5, "time in seconds that the resources will be re synced")
	// TODO: Deal with config file after this initial attempt.
	pflag.String("configFile", "", "config file that should should be used. The config will override all other command line values")
	pflag.String("api-version", "", "Kubernetes apiVersion and has a format of $GROUP_NAME/$VERSION (e.g app.example.com/v1alpha1)")
	pflag.String("kind", "", "Kubernetes CustomResourceDefintion kind. (e.g AppService)")
	pflag.String("apb-image", "", "Apb Image path for which we can get the APB spec.")
	pflag.String("plan", "", "Plan that the operator should be interacting with for an APB.")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func main() {
	printVersion()

	resource := viper.GetString("api-version")
	kind := viper.GetString("kind")
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Fatalf("Failed to get watch namespace: %v", err)
	}
	resyncPeriod := viper.GetInt("resync")
	logrus.Infof("Watching %s, %s, %s, %d", resource, kind, namespace, resyncPeriod)
	spec := getSpec(viper.GetString("apb-image"))
	//register type with gvk
	registerType(resource, kind)

	sdk.Watch(resource, kind, namespace, resyncPeriod)
	plan, ok := spec.GetPlan(viper.GetString("plan"))
	if !ok {
		logrus.Errorf("unable to find plan: %v in the spec for apb-image: %v", viper.GetString("plan"), viper.GetString("apb-image"))
		os.Exit(2)
	}
	sdk.Handle(stub.NewHandler(map[string]crd.SpecPlan{
		fmt.Sprintf("%v:%v", resource, kind): crd.SpecPlan{
			Spec: spec,
			Plan: plan,
		},
	}))
	sdk.Run(context.TODO())
}

func registerType(resource, kind string) {
	gv := strings.Split(resource, "/")
	schemeGroupVersion := schema.GroupVersion{Group: gv[0], Version: gv[1]}
	schemeBuilder := k8sruntime.NewSchemeBuilder(func(s *k8sruntime.Scheme) error {
		s.AddKnownTypeWithName(schema.GroupVersionKind{
			Group:   gv[0],
			Version: gv[1],
			Kind:    kind,
		}, &unstructured.Unstructured{})
		metav1.AddToGroupVersion(s, schemeGroupVersion)
		return nil
	})
	k8sutil.AddToSDKScheme(schemeBuilder.AddToScheme)
}

func getSpec(image string) bundle.Spec {
	spec := `
version: 1.0
name: postgresql-apb
description: SCL PostgreSQL apb implementation
bindable: true
async: optional
tags:
  - database
  - postgresql
metadata:
  documentationUrl: https://www.postgresql.org/docs/
  longDescription: An apb that deploys postgresql 9.4, 9.5, or 9.6.
  dependencies:
    - 'registry.access.redhat.com/rhscl/postgresql-94-rhel7'
    - 'registry.access.redhat.com/rhscl/postgresql-95-rhel7'
    - 'registry.access.redhat.com/rhscl/postgresql-96-rhel7'
  displayName: PostgreSQL (APB)
  console.openshift.io/iconClass: icon-postgresql
  providerDisplayName: "Red Hat, Inc."
plans:
  - name: dev
    description: A single DB server with no storage
    free: true
    metadata:
      displayName: Development
      longDescription:
        This plan provides a single non-HA PostgreSQL server without
        persistent storage
      cost: $0.00
    parameters:
      - name: postgresql_database
        default: admin
        type: string
        title: PostgreSQL Database Name
        pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$"
        required: true
      - name: postgresql_user
        default: admin
        title: PostgreSQL User
        type: string
        maxlength: 63
        pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$"
        required: true
      - name: postgresql_password
        type: string
        title: PostgreSQL Password
        display_type: password
        pattern: "^[a-zA-Z0-9_~!@#$%^&*()-=<>,.?;:|]+$"
        required: true
      - name: postgresql_version
        default: '9.6'
        enum: ['9.6', '9.5', '9.4']
        type: enum
        title: PostgreSQL Version
        required: true
        updatable: true
    updates_to:
      - prod
  - name: prod
    description: A single DB server with persistent storage
    free: true
    metadata:
      displayName: Production
      longDescription:
        This plan provides a single non-HA PostgreSQL server with
        persistent storage
      cost: $0.00
    parameters:
      - name: postgresql_database
        default: admin
        type: string
        title: PostgreSQL Database Name
        pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$"
        required: true
      - name: postgresql_user
        default: admin
        title: PostgreSQL User
        type: string
        maxlength: 63
        pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$"
        required: true
      - name: postgresql_password
        type: string
        title: PostgreSQL Password
        display_type: password
        pattern: "^[a-zA-Z0-9_~!@#$%^&*()-=<>,.?;:|]+$"
        required: true
      - name: postgresql_version
        default: '9.6'
        enum: ['9.6', '9.5', '9.4']
        type: enum
        title: PostgreSQL Version
        required: true
        updatable: true
      - name: postgresql_volume_size
        type: enum
        default: '1Gi'
        enum: ['1Gi', '5Gi', '10Gi']
        title: PostgreSQL Volume Size
        required: true
    updates_to:
      - dev
`

	s := bundle.Spec{}
	err := yaml.Unmarshal([]byte(spec), &s)
	if err != nil {
		logrus.Infof("unable to get the spec - %v", err)
	}
	return s
}
