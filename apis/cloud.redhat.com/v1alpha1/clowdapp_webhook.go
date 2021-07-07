/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var clowdapplog = logf.Log.WithName("clowdapp-resource")

func (r *ClowdApp) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-cloud-redhat-com-v1alpha1-clowdapp,mutating=false,failurePolicy=fail,sideEffects=None,groups=cloud.redhat.com,resources=clowdapps,verbs=create;update,versions=v1alpha1,name=vclowdapp.kb.io,admissionReviewVersions={v1,v1beta1,v1alpha1}

var _ webhook.Validator = &ClowdApp{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ClowdApp) ValidateCreate() error {
	clowdapplog.Info("validate create", "name", r.Name)

	return r.processValidations(r,
		validateDatabase,
		validateSidecars,
	)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ClowdApp) ValidateUpdate(old runtime.Object) error {
	clowdapplog.Info("validate update", "name", r.Name)

	return r.processValidations(r,
		validateDatabase,
		validateSidecars,
	)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ClowdApp) ValidateDelete() error {
	clowdapplog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

type appValidationFunc func(*ClowdApp) field.ErrorList

func (r *ClowdApp) processValidations(o *ClowdApp, vfns ...appValidationFunc) error {
	var allErrs field.ErrorList

	for _, validation := range vfns {
		fieldList := validation(o)
		if fieldList != nil {
			allErrs = append(allErrs, fieldList...)
		}
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "cloud.redhat.com", Kind: "ClowdApp"},
		r.Name, allErrs,
	)
}

func validateDatabase(r *ClowdApp) field.ErrorList {
	if r.Spec.Database.Name != "" && r.Spec.Database.SharedDBAppName != "" {
		return field.ErrorList{field.Forbidden(
			field.NewPath("spec.Database.Name", "spec.Database.SharedDBAppName"), "cannot set db name and sharedDbApp Name together"),
		}
	}
	return nil
}

func validateSidecars(r *ClowdApp) field.ErrorList {
	allErrs := field.ErrorList{}
	for depIndx, deployment := range r.Spec.Deployments {
		for carIndx, sidecar := range deployment.PodSpec.Sidecars {
			if sidecar.Name != "token-refresher" {
				allErrs = append(
					allErrs,
					field.Forbidden(
						field.NewPath(fmt.Sprintf("spec.Deployment[%d].Sidecars[%d]", depIndx, carIndx)),
						"Sidecar is of unknown type, must be one of [token-refresher]",
					),
				)
			}
		}
	}
	for jobIndx, job := range r.Spec.Jobs {
		if job.Schedule == "" {
			continue
		}
		for carIndx, sidecar := range job.PodSpec.Sidecars {
			if sidecar.Name != "token-refresher" {
				allErrs = append(
					allErrs,
					field.Forbidden(
						field.NewPath(fmt.Sprintf("spec.Deployment[%d].Sidecars[%d]", jobIndx, carIndx)),
						"Sidecar is of unknown type, must be one of [token-refresher]",
					),
				)
			}
		}
	}
	return allErrs
}
