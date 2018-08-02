package stub

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/philbrookes/phils-operator/pkg/apis/app/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"

	ptv1a1 "github.com/philbrookes/phils-operator/pkg/apis/app/v1alpha1"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewHandler(svcClient *sc.Clientset) sdk.Handler {
	return &Handler{svcClient: svcClient}
}

type Handler struct {
	svcClient *sc.Clientset
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch philsThing := event.Object.(type) {
	case *v1alpha1.PhilsThing:
		if !event.Deleted {
			return h.handlePhilsThing(philsThing)
		}
		return h.handlePhilsThingDelete(philsThing)
	}
	return nil
}

func (h *Handler) handlePhilsThingDelete(philsThing *ptv1a1.PhilsThing) error {
	return h.svcClient.Servicecatalog().ServiceInstances(philsThing.Namespace).Delete(philsThing.Spec.ServiceInstanceName, &metav1.DeleteOptions{})
}

func (h *Handler) handlePhilsThing(philsThing *ptv1a1.PhilsThing) error {
	philsThingCopy := philsThing.DeepCopy()
	switch philsThingCopy.Status.Phase {
	case "":
		philsThingCopy.Status.Phase = "accepted"
	case "accepted":
		// do the provision
		scs, err := h.svcClient.Servicecatalog().ClusterServiceClasses().List(metav1.ListOptions{})
		if err != nil {
			return err
		}

		serviceClass := v1beta1.ClusterServiceClass{}

		for _, sc := range scs.Items {
			if sc.Spec.CommonServiceClassSpec.ExternalName == philsThingCopy.Spec.ServiceClassName {
				serviceClass = sc
			}
		}

		// would be better to get serviceClass from a function and return an error if not found.
		// testing this here for the sake of squeezing this into 1-hour
		if serviceClass.GetName() == "" {
			philsThingCopy.Status.Phase = "failed"
			return errors.New("Could not find service class: " + philsThingCopy.Spec.ServiceClassName)
		}

		parameters, err := json.Marshal(philsThingCopy.Spec.Params)
		if err != nil {
			return err
		}

		serviceInstance := v1beta1.ServiceInstance{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "servicecatalog.k8s.io/v1beta1",
				Kind:       "ServiceInstance",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    philsThingCopy.GetNamespace(),
				GenerateName: philsThingCopy.GetName() + "-",
			},
			Spec: v1beta1.ServiceInstanceSpec{
				PlanReference: v1beta1.PlanReference{
					ClusterServiceClassExternalName: serviceClass.Spec.ExternalName,
				},
				ClusterServiceClassRef: &v1beta1.ClusterObjectReference{
					Name: serviceClass.Name,
				},
				ClusterServicePlanRef: &v1beta1.ClusterObjectReference{
					Name: "default",
				},
				Parameters: &runtime.RawExtension{Raw: parameters},
			},
		}

		si, err := h.svcClient.Servicecatalog().ServiceInstances(philsThingCopy.GetNamespace()).Create(&serviceInstance)
		if err != nil {
			return err
		}
		philsThingCopy.Spec.ServiceInstanceName = si.GetName()
		philsThingCopy.Status.Phase = "provisioning"

	case "provisioning":
		// check the progress
		si, err := h.svcClient.ServicecatalogV1beta1().ServiceInstances(philsThingCopy.Namespace).Get(philsThingCopy.Spec.ServiceInstanceName, metav1.GetOptions{})
		if err != nil {
			return nil
		}
		//if finished: philsThingCopy.Status.Phase "complete"
		for _, cnd := range si.Status.Conditions {
			if cnd.Type == "Ready" && cnd.Status == "True" {
				philsThingCopy.Status.Phase = "complete"
			}
		}
	}

	return sdk.Update(philsThingCopy)
}
