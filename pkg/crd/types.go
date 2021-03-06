package crd

import (
	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/pborman/uuid"
)

// BundlePhase - Current phase of the bundle.
type BundlePhase string

const (
	//BundlePhaseInit - inital phase
	BundlePhaseInit BundlePhase = ""
	//BundlePhaseCreating - creating phase
	BundlePhaseCreating BundlePhase = "creating"
	//BundlePhaseRunning - running phase
	BundlePhaseRunning BundlePhase = "running"
	//BundlePhaseDeleting - deleting phase
	BundlePhaseDeleting BundlePhase = "deleting"
	//BundlePhaseFailed - failed phase
	BundlePhaseFailed BundlePhase = "failed"
)

// BundleStatus - generic bundle status
type BundleStatus struct {
	Phase          BundlePhase            `json:"phase"`
	Message        string                 `json:"message"`
	Parameters     string                 `json:"parameters"`
	ID             *uuid.UUID             `json:"id,omitempty"`
	StatusMessages []bundle.StatusMessage `json:"statusMessages"`
}

// SpecPlan - the spec and plan for a GVK
type SpecPlan struct {
	Spec bundle.Spec
	Plan bundle.Plan
}
