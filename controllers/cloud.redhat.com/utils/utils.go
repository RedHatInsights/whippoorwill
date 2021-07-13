package utils

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/errors"
	obj "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/object"
	"github.com/go-logr/logr"
	core "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

const rCharSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Log is a null logger instance.
var Log logr.Logger = ctrllog.NullLogger{}

// RandString generates a random string of length n
func RandString(n int) string {
	b := make([]byte, n)

	for i := range b {
		b[i] = rCharSet[rand.Intn(len(rCharSet))]
	}

	return string(b)
}

// Updater is a bool type object with functions attached that control when a resource should be
// created or applied.
type Updater bool

// Apply will apply the resource if it already exists, and create it if it does not. This is based
// on the bool value of the Update object.
func (u *Updater) Apply(ctx context.Context, cl client.Client, obj client.Object) error {
	var err error
	var kind string

	if obj.GetObjectKind().GroupVersionKind().Kind == "" {
		kind = reflect.TypeOf(obj).String()
	} else {
		kind = obj.GetObjectKind().GroupVersionKind().Kind
	}

	meta := obj.(metav1.Object)

	if *u {
		//Log.Info("Updating resource", "namespace", meta.GetNamespace(), "name", meta.GetName(), "kind", kind)
		err = cl.Update(ctx, obj)
	} else {
		if meta.GetName() == "" {
			Log.Info("Skipping resource as name unknown", "kind", kind)
			return nil
		}

		//Log.Info("Creating resource", "namespace", meta.GetNamespace(), "name", meta.GetName(), "kind", kind)
		err = cl.Create(ctx, obj)
	}

	if err != nil {
		verb := "creating"
		if *u {
			verb = "updating"
		}

		msg := fmt.Sprintf("Error %s resource %s %s", verb, kind, meta.GetName())
		return errors.Wrap(msg, err)
	}

	return nil
}

// UpdateOrErr returns an update object if the err supplied is nil.
func UpdateOrErr(err error) (Updater, error) {
	update := Updater(err == nil)

	if err != nil && !k8serr.IsNotFound(err) {
		return update, errors.Wrap("Failed to update", err)
	}

	return update, nil
}

// UpdateAllOrErr queries the client for a range of objects and returns updater objects for each.
func UpdateAllOrErr(ctx context.Context, cl client.Client, nn types.NamespacedName, obj ...client.Object) (map[client.Object]Updater, error) {
	updates := map[client.Object]Updater{}

	for _, resource := range obj {
		update, err := UpdateOrErr(cl.Get(ctx, nn, resource))

		if err != nil {
			return updates, err
		}

		updates[resource] = update
	}

	return updates, nil
}

// ApplyAll applies all the update objects in the list called updates.
func ApplyAll(ctx context.Context, cl client.Client, updates map[client.Object]Updater) error {
	for resource, update := range updates {
		if err := update.Apply(ctx, cl, resource); err != nil {
			return err
		}
	}

	return nil
}

// B64Decode decodes the provided secret
func B64Decode(s *core.Secret, key string) (string, error) {
	decoded, err := b64.StdEncoding.DecodeString(string(s.Data[key]))

	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

func intMinMax(listStrInts []string, max bool) (string, error) {
	var listInts []int
	for _, strint := range listStrInts {
		i, err := strconv.Atoi(strint)
		if err != nil {
			return "", errors.Wrap("Failed to convert", err)
		}
		listInts = append(listInts, i)
	}
	ol := listInts[0]
	for i, e := range listInts {
		if max {
			if i == 0 || e > ol {
				ol = e
			}
		} else {
			if i == 0 || e < ol {
				ol = e
			}
		}
	}
	return strconv.Itoa(ol), nil
}

// IntMin takes a list of integers as strings and returns the minimum.
func IntMin(listStrInts []string) (string, error) {
	return intMinMax(listStrInts, false)
}

// IntMax takes a list of integers as strings and returns the maximum.
func IntMax(listStrInts []string) (string, error) {
	return intMinMax(listStrInts, true)
}

// ListMerge takes a list comma separated strings and performs a set union on them.
func ListMerge(listStrs []string) (string, error) {
	optionStrings := make(map[string]bool)
	for _, optionsList := range listStrs {
		brokenString := strings.Split(optionsList, ",")
		for _, option := range brokenString {
			optionStrings[strings.TrimSpace(option)] = true
		}
	}
	keys := make([]string, len(optionStrings))

	i := 0
	for key := range optionStrings {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return strings.Join(keys, ","), nil
}

// MakeLabeler creates a function that will label objects with metadata from
// the given namespaced name and labels
func MakeLabeler(nn types.NamespacedName, labels map[string]string, obj obj.LabeledClowdObject) func(metav1.Object) {
	return func(o metav1.Object) {
		o.SetName(nn.Name)
		o.SetNamespace(nn.Namespace)
		o.SetLabels(labels)
		o.SetOwnerReferences([]metav1.OwnerReference{obj.MakeOwnerReference()})
	}
}

// GetCustomLabeler takes a set of labels and returns a labeler function that
// will apply those labels to a reource.
func GetCustomLabeler(labels map[string]string, nn types.NamespacedName, baseResource obj.LabeledClowdObject) func(metav1.Object) {
	appliedLabels := baseResource.GetLabels()
	for k, v := range labels {
		appliedLabels[k] = v
	}
	return MakeLabeler(nn, appliedLabels, baseResource)
}

// MakeService takes a service object and applies the correct ownership and labels to it.
func MakeService(service *core.Service, nn types.NamespacedName, labels map[string]string, ports []core.ServicePort, baseResource obj.ClowdObject, nodePort bool) {
	labeler := GetCustomLabeler(labels, nn, baseResource)
	labeler(service)
	service.Spec.Selector = labels
	service.Spec.Ports = ports
	if nodePort {
		service.Spec.Type = "NodePort"
	} else {
		service.Spec.Type = "ClusterIP"
	}
}

// MakePVC takes a PVC object and applies the correct ownership and labels to it.
func MakePVC(pvc *core.PersistentVolumeClaim, nn types.NamespacedName, labels map[string]string, size string, baseResource obj.ClowdObject) {
	labeler := GetCustomLabeler(labels, nn, baseResource)
	labeler(pvc)
	pvc.Spec.AccessModes = []core.PersistentVolumeAccessMode{core.ReadWriteOnce}
	pvc.Spec.Resources = core.ResourceRequirements{
		Requests: core.ResourceList{
			core.ResourceName(core.ResourceStorage): resource.MustParse(size),
		},
	}
}

// IntPtr returns a pointer to the passed integer.
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to the passed boolean.
func BoolPtr(b bool) *bool {
	boolVar := b
	return &boolVar
}

// GetKindFromObj retrieves GVK associated with registered runtime.Object
func GetKindFromObj(scheme *runtime.Scheme, object runtime.Object) (schema.GroupVersionKind, error) {
	gvks, nok, err := scheme.ObjectKinds(object)

	if err != nil {
		return schema.EmptyObjectKind.GroupVersionKind(), err
	}

	if nok {
		return schema.EmptyObjectKind.GroupVersionKind(), fmt.Errorf("object type is unknown")
	}

	return gvks[0], nil
}

func GetClowderNamespace() (string, error) {
	clowderNsB, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")

	// CLOBBER the error here as this is our default
	if err != nil {
		return "clowder-system", nil
	}

	return string(clowderNsB), nil
}

// CopySecret will return a *core.Secret that is copied from a source NamespaceName and intended to
// be applied into a destination NamespacedName
func CopySecret(ctx context.Context, client client.Client, srcSecretRef types.NamespacedName, dstSecretRef types.NamespacedName) (error, *core.Secret) {
	nullRef := types.NamespacedName{}
	if srcSecretRef == nullRef {
		return errors.New("srcSecretRef is an empty NamespacedName"), nil
	}
	if dstSecretRef == nullRef {
		return errors.New("dstSecretRef is an empty NamespacedName"), nil
	}

	srcSecret := &core.Secret{}

	if err := client.Get(ctx, srcSecretRef, srcSecret); err != nil {
		return err, nil
	}

	newSecret := &core.Secret{}
	newSecret.Immutable = srcSecret.Immutable
	newSecret.Data = srcSecret.Data
	newSecret.Type = srcSecret.Type
	newSecret.SetName(dstSecretRef.Name)
	newSecret.SetNamespace(dstSecretRef.Namespace)

	return nil, newSecret
}
