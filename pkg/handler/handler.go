package stub

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"

	"github.com/automationbroker/automation-operator/pkg/crd"
	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/runtime"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	statusKey = "status"
	specKey   = "spec"
)

// NewHandler - Creates a new handler
func NewHandler(b map[string]crd.SpecPlan) sdk.Handler {
	//Setting the runtime for bundle lib.
	runtime.NewRuntime(runtime.Configuration{})
	return &Handler{
		bundles: b,
	}
}

// Handler - to handle bundle CRs
type Handler struct {
	// all of the gvk to bundle objects the operator is handeling.
	bundles map[string]crd.SpecPlan
}

// Handle - handle the events
func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	if s, ok := h.bundles[getGVKString(event)]; ok {
		// Should be unstructured type and we should be able to get the "spec" for the parameters
		// we should unmarshall the "status" section, if available to the status section
		logrus.Infof("here - %#v", event.Object)
		// Need to get json schema
		handleSpecEvent(s, event)
	}
	return nil
}

func getGVKString(e sdk.Event) string {
	return fmt.Sprintf("%v/%v:%v",
		e.Object.GetObjectKind().GroupVersionKind().Group,
		e.Object.GetObjectKind().GroupVersionKind().Version,
		e.Object.GetObjectKind().GroupVersionKind().Kind)
}

func handleSpecEvent(s crd.SpecPlan, event sdk.Event) {
	o, ok := event.Object.(*unstructured.Unstructured)
	if !ok {
		logrus.Infof("unable to create unstructured object from event object")
		return
	}
	crStatus := crd.BundleStatus{}
	// Using status as a hardcoded value TODO: make const?
	status, ok := o.UnstructuredContent()[statusKey]
	if ok {
		i, err := json.Marshal(status)
		if err != nil {
			logrus.Errorf("unable to create json from the map - %v", err)
		}
		err = json.Unmarshal(i, &crStatus)
		if err != nil {
			logrus.Errorf("unable to create json from the map - %v", err)
		}
	}

	specMap := map[string]interface{}{}
	if spec, ok := o.UnstructuredContent()[specKey]; ok {
		sp, ok := spec.(map[string]interface{})
		if !ok {
			logrus.Infof("unable to deal with object with incorrect spec field: %v", o.GetName())
			// Save Error State and message on the CR.
			crStatus.Message = "unable to understand spec field"
			crStatus.Phase = crd.BundlePhaseFailed
			o.UnstructuredContent()[statusKey] = crStatus
			sdk.Update(o)
			return
		}
		specMap = sp
	}

	if crStatus.Phase == crd.BundlePhaseInit || crStatus.Phase == crd.BundlePhaseFailed {
		// Here we need to default the parameters based on the json spec from the parameters.
		// We should update the crStatus Phase to creating and save.
		changed := false
		for _, p := range s.Plan.Parameters {
			v, ok := specMap[p.Name]
			switch {
			case !ok && p.Default != nil:
				specMap[p.Name] = p.Default
				changed = true
				continue
			case !ok && p.Required:
				logrus.Infof("unable to deal with parameter w/o default but required resource: %v, Parameter: %v needs a value", o.GetName(), p.Name)
				// Save Error State and message on the CR.
				//TODO: this should gather all the errors and then have all the errors reported in the message.
				crStatus.Message = fmt.Sprintf("Paramater: %v must have a value", p.Name)
				crStatus.Phase = crd.BundlePhaseFailed
				o.UnstructuredContent()[statusKey] = crStatus
				sdk.Update(o)
				return
			case !ok && p.Default == nil:
				continue
			case ok:
				if ok, err := validParameter(p, v); !ok {
					logrus.Infof("unable to deal with parameter w/o default but required resource: %v, Parameter: %v needs a value", o.GetName(), p.Name)
					// Save Error State and message on the CR.
					//TODO: this should gather all the errors and then have all the errors reported in the message.
					crStatus.Message = fmt.Sprintf("Paramater: %v invalide: %v", p.Name, err)
					crStatus.Phase = crd.BundlePhaseFailed
					o.UnstructuredContent()[statusKey] = crStatus
					sdk.Update(o)
					return
				}
			default:
				logrus.Errorf("unknown condition of parameter and spec value: default: %v, required: %v, found: %v", p.Default, p.Required, ok)
			}
		}
		if changed {
			//Update CR with th defaults. We will return and wait for the update event to re-trigger the rest of this.
			crStatus.Phase = crd.BundlePhaseCreating
			crStatus.Message = ""
			o.UnstructuredContent()[statusKey] = crStatus
			o.UnstructuredContent()[specKey] = specMap
			sdk.Update(o)
			return
		}
	}
	// Core Reconcile Loop will happen here

	// Hash the paramters
	specHash, err := hashMap(specMap)
	if err != nil {
		logrus.Infof("unable to hash spec parameters: %v", err)
		crStatus.Message = fmt.Sprintf("could not hash parameters")
		crStatus.Phase = crd.BundlePhaseFailed
		o.UnstructuredContent()[statusKey] = crStatus
		sdk.Update(o)
		return
	}
	if crStatus.ID == nil {
		// Generate ID.
		id := uuid.NewRandom()
		crStatus.ID = &id
	}
	if crStatus.Parameters != specHash {
		crStatus.Parameters = specHash
		err := launchAPBProvision(specMap, s, &crStatus, o)
		if err != nil {
			logrus.Infof("unable to launch apb: %v", err)
			crStatus.Message = fmt.Sprintf("could not launch apb")
			crStatus.Phase = crd.BundlePhaseFailed
			o.UnstructuredContent()[statusKey] = crStatus
			sdk.Update(o)
			return
		}
		crStatus.Phase = crd.BundlePhaseCreating
		o.UnstructuredContent()[statusKey] = crStatus
		sdk.Update(o)
		return
	}

}

//TODO: Actaully write this logic here.
func validParameter(b bundle.ParameterDescriptor, value interface{}) (bool, error) {
	return true, nil
}

func hashMap(m map[string]interface{}) (string, error) {
	h := sha1.New()
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func launchAPBProvision(specMap map[string]interface{}, specPlan crd.SpecPlan, status *crd.BundleStatus, o *unstructured.Unstructured) error {
	specMap["_apb_plan_id"] = specPlan.Plan.ID
	p := bundle.Parameters(specMap)
	si := bundle.ServiceInstance{
		ID:   *status.ID,
		Spec: &specPlan.Spec,
		Context: &bundle.Context{
			Platform:  "kubernetes",
			Namespace: o.GetNamespace(),
		},
		Parameters: &p,
	}
	logrus.Infof("using service instance : %v", si)
	ex := bundle.NewExecutor()
	channel := ex.Provision(&si)
	go func() {
		for status := range channel {
			logrus.Infof("status: %v", status)
		}
	}()
	return nil
}
