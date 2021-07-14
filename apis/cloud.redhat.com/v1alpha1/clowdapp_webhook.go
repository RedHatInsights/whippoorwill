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
//+kubebuilder:webhook:path=/mutate-pod,mutating=true,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create;update,versions=v1,name=vclowdapp.kb.io,admissionReviewVersions={v1,v1beta1,v1alpha1}

var _ webhook.Validator = &ClowdApp{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ClowdApp) ValidateCreate() error {
	clowdapplog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList

	if r.Spec.Database.Name != "" && r.Spec.Database.SharedDBAppName != "" {
		allErrs = append(allErrs, field.Forbidden(
			field.NewPath("spec.Database.Name", "spec.Database.SharedDBAppName"), "cannot set db name and sharedDbApp Name together"),
		)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "cloud.redhat.com", Kind: "ClowdApp"},
		r.Name, allErrs,
	)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ClowdApp) ValidateUpdate(old runtime.Object) error {
	clowdapplog.Info("validate update", "name", r.Name)

	var allErrs field.ErrorList

	if r.Spec.Database.Name != "" && r.Spec.Database.SharedDBAppName != "" {
		allErrs = append(allErrs, field.Forbidden(
			field.NewPath("spec.Database.Name", "spec.Database.SharedDBAppName"), "cannot set db name and sharedDbApp Name together"),
		)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "cloud.redhat.com", Kind: "ClowdApp"},
		r.Name, allErrs,
	)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ClowdApp) ValidateDelete() error {
	clowdapplog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
