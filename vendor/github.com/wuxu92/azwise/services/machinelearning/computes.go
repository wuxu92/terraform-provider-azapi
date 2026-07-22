// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package machinelearning

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Computes provides resource knowledge for
// Microsoft.MachineLearningServices/workspaces/computes.
//
// One ARM type, four AzureRM resources discriminated by `properties.computeType`
// (see machinelearningcomputes.ComputeResource.Properties, a discriminated union):
//   - azurerm_machine_learning_compute_cluster    (computeType = "AmlCompute")
//   - azurerm_machine_learning_compute_instance    (computeType = "ComputeInstance")
//   - azurerm_machine_learning_inference_cluster    (computeType = "AKS")
//   - azurerm_machine_learning_synapse_spark        (computeType = "SynapseSpark")
//
// Kind-specific properties live under `properties.properties.*` (e.g. AmlCompute's
// vmSize/vmPriority/scaleSettings, ComputeInstance's vmSize/subnet). Only universal or
// per-kind-safe knowledge is encoded:
//   - Every contributing resource marks `name` ForceNew and the parent
//     machine_learning_workspace_id ForceNew (envelope/parent — excluded from body rules).
//   - Value constraints on a sub-object that only one kind sets are safe because they
//     fire only when that path is present: the AmlCompute vmPriority enum below only
//     validates a compute_cluster body.
//
// NOTE: `properties.computeType` is the discriminator; changing it replaces the resource,
// but it is not user-editable in a stable body, so it is not listed as a ForceNew rule.
// Name validators differ per kind (compute_cluster: 3-32 start-with-letter; datastore/
// instance differ) so no universal name StringRule is emitted.
//
// Sources:
//   - AzureRM internal/services/machinelearning/machine_learning_compute_cluster_resource.go
//     (schema :26-165, Create :167-262, vmPriority enum :71)
//   - AzureRM internal/services/machinelearning/machine_learning_compute_instance_resource.go
//     (ComputeInstance body :232)
//   - AzureRM internal/services/machinelearning/machine_learning_inference_cluster_resource.go
//     (AKS body :294)
//   - AzureRM internal/services/machinelearning/machine_learning_synapse_spark_resource.go
//     (SynapseSpark body :119)
//   - AzureRM internal/services/machinelearning/validate/compute_cluster_name.go
//   - go-azure-sdk .../machinelearningservices/2025-06-01/machinelearningcomputes:
//     model_computeresource.go, model_compute.go, model_amlcompute.go,
//     model_amlcomputeproperties.go, constants.go (ComputeType, VMPriority)
type Computes struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Computes)(nil)

func NewComputes() *Computes {
	return &Computes{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.MachineLearningServices/workspaces/computes",
			ApiVersions:  []string{"2025-06-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute, // only compute_cluster supports update; others recreate
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// compute_cluster vm_priority (AmlCompute only). Only fires when the
					// AmlComputeProperties.vmPriority path is present. Full ARM SDK set.
					PropertyPath:  "properties.properties.vmPriority",
					AllowedValues: []string{"Dedicated", "LowPriority"},
					Message:       "must be one of Dedicated or LowPriority",
				},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewComputes()) }
