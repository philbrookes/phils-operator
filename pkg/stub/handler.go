package stub

import (
	"context"
	"fmt"

	"github.com/philbrookes/phils-operator/pkg/apis/app/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"

	ptv1a1 "github.com/philbrookes/phils-operator/pkg/apis/app/v1alpha1"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch philsThing := event.Object.(type) {
	case *v1alpha1.PhilsThing:
		if !event.Deleted {
			return h.handlePhilsThing(philsThing)
		}
	}
	return nil
}

func (h *Handler) handlePhilsThing(philsThing *ptv1a1.PhilsThing) error {
	philsThingCopy := philsThing.DeepCopy()

	switch philsThingCopy.Spec.PhilsData {
	case "complete":
		fmt.Printf("process complete!\n")
		return sdk.Delete(philsThingCopy)
	case "provision_required":
		fmt.Printf("on stage1, count: %v\n", philsThingCopy.Spec.PhilsCounter)
		if len(philsThingCopy.Spec.PhilsCounter) < 10 {
			philsThingCopy.Spec.PhilsCounter = philsThingCopy.Spec.PhilsCounter + "|"
		} else {
			fmt.Printf("going to stage 2\n")
			philsThingCopy.Spec.PhilsCounter = ""
			philsThingCopy.Spec.PhilsData = "provision_in_progress"
		}
	case "provision_in_progress":
		fmt.Printf("on stage2, count: %v\n", philsThingCopy.Spec.PhilsCounter)
		if len(philsThingCopy.Spec.PhilsCounter) < 10 {
			philsThingCopy.Spec.PhilsCounter = philsThingCopy.Spec.PhilsCounter + "|"
		} else {
			fmt.Printf("going to complete\n")
			philsThingCopy.Spec.PhilsCounter = ""
			philsThingCopy.Spec.PhilsData = "complete"
		}
	default:
		fmt.Printf("unknown data, defaulting to stage1\n")
		philsThingCopy.Spec.PhilsData = "provision_required"
	}
	return sdk.Update(philsThingCopy)
}
