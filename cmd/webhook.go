package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	v1 "k8s.io/kubernetes/pkg/apis/core/v1"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

var (
	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}

	// Extra labels to add
	addLabels = map[string]string{}
)

const (
	admissionWebhookAnnotationValidateKey = "ip.brivo.com/validate"
	// admissionWebhookAnnotationMutateKey   = "ip.brivo.com/mutate"
	admissionWebhookAnnotationStatusKey = "ip.brivo.com/status"

	brivoIPLabel     = "ip.brivo.com/address" // IP Address(es)
	brivoIPRange     = "ip.brivo.com/range"
	awsEIPAnnotation = "service.beta.kubernetes.io/aws-load-balancer-eip-allocations"
)

type WebhookServer struct {
	server *http.Server
}

// WhSvrParameters : Webhook Server parameters
type WhSvrParameters struct {
	port           int    // webhook server port
	certFile       string // path to the x509 certificate for https
	keyFile        string // path to the x509 private key matching `CertFile`
	sidecarCfgFile string // path to sidecar injector configuration file
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	// defaulting with webhooks:
	// https://github.com/kubernetes/kubernetes/issues/57982
	_ = v1.AddToScheme(runtimeScheme)
}

// Check if we need to mutate
func mutationRequired(ignoredList []string, metadata *metav1.ObjectMeta) bool {
	required := admissionRequired(ignoredList, brivoIPLabel, metadata)
	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	// If we don't have the brivo IP annotation, we don't need to mutate this.
	if _, ok := annotations[brivoIPLabel]; ok {
		glog.Info("Found brivo IP annotation.  We should modify this object.")
		required = true
	} else {
		glog.Info("Did not find brivo IP annotation.  We should NOT modify this object.")
		required = false
	}
	status := annotations[admissionWebhookAnnotationStatusKey]

	if strings.ToLower(status) == "mutated" {
		glog.Info("We have already mutated this object (found key status=mutated).  Skipping.")
		required = false
	}

	glog.Infof("Mutation policy for %v/%v: required:%v", metadata.Namespace, metadata.Name, required)
	return required
}

func updateAnnotation(target map[string]string, added map[string]string) (patch []patchOperation) {
	var ipstr string

	if ip, found := target[brivoIPLabel]; found {
		glog.Info("Found annotation " + brivoIPLabel)
		if strings.Contains(ip, "/") {
			glog.Info("Is this a CIDR? I don't know what to do with these yet... ")
		}
		// sess := session.Must(session.NewSessionWithOptions(session.Options{
		// 	SharedConfigState: session.SharedConfigEnable,
		// }))
		aresult, aerr := GetAddressOrAllocate(strings.Split(ip, ","))
		if aerr != nil {
			glog.Error("Got an error retrieving the Elastic IP addresses")
			glog.Error(aerr)
			// Set status key to failed
			added[admissionWebhookAnnotationStatusKey] = "failed"
		}
		ipstr = ""
		for _, addr := range aresult.Addresses {
			glog.Infof("IP address:    %v", *addr.PublicIp)
			glog.Infof("Allocation ID: %v", *addr.AllocationId)
			ipstr += *addr.AllocationId
			ipstr += ","
		}
		added[awsEIPAnnotation] = ipstr
	}
	glog.Infof("target: %v", target)
	glog.Infof("added:  %v", added)

	for key, value := range added {
		if target[key] != "" {
			patch = append(patch, patchOperation{
				Op:    "replace",
				Path:  "/metadata/annotations/" + key,
				Value: value,
			})
		} else {
			// "~"(tilde) is encoded as "~0" and "/"(forward slash) is encoded as "~1".
			patch = append(patch, patchOperation{
				Op:    "add",
				Path:  "/metadata/annotations/" + strings.ReplaceAll(key, "/", "~1"),
				Value: value,
			})
		}
	}

	// values := make(map[string]string)
	//
	// // Find values that have not been added to the map
	// for key, value := range added {
	// 	if target == nil || target[key] == "" {
	// 		values[key] = value
	// 	}
	// }

	return patch
}

func createPatch(availableAnnotations map[string]string, annotations map[string]string, availableLabels map[string]string, labels map[string]string) ([]byte, error) {
	var patch []patchOperation

	patch = append(patch, updateAnnotation(availableAnnotations, annotations)...)
	// patch = append(patch, updateLabels(availableLabels, labels)...)

	return json.Marshal(patch)
}

// func createIPatch()

// main mutation process
func (whsvr *WebhookServer) mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	var (
		availableLabels, availableAnnotations map[string]string
		objectMeta                            *metav1.ObjectMeta
		resourceNamespace, resourceName       string
	)

	glog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, resourceName, req.UID, req.Operation, req.UserInfo)

	switch req.Kind.Kind {
	// case "Deployment":
	// 	var deployment appsv1.Deployment
	// 	if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
	// 		glog.Errorf("Could not unmarshal raw object: %v", err)
	// 		return &v1beta1.AdmissionResponse{
	// 			Result: &metav1.Status{
	// 				Message: err.Error(),
	// 			},
	// 		}
	// 	}
	// 	resourceName, resourceNamespace, objectMeta = deployment.Name, deployment.Namespace, &deployment.ObjectMeta
	// 	availableAnnotations = deployment.Annotations
	// 	availableLabels = deployment.Labels
	case "Service":
		var service corev1.Service
		if err := json.Unmarshal(req.Object.Raw, &service); err != nil {
			glog.Errorf("Could not unmarshal raw object: %v", err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		resourceName, resourceNamespace, objectMeta = service.Name, service.Namespace, &service.ObjectMeta
		availableLabels = service.Labels
		availableAnnotations = service.Annotations
	}

	if !mutationRequired(ignoredNamespaces, objectMeta) {
		glog.Infof("Skipping validation for %s/%s due to policy check", resourceNamespace, resourceName)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	annotations := map[string]string{admissionWebhookAnnotationStatusKey: "mutated"}
	// available: labels on deploy/service

	patchBytes, err := createPatch(availableAnnotations, annotations, availableLabels, addLabels)
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	glog.Infof("AdmissionResponse: patch=%v\n", string(patchBytes))
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

// Serve method for webhook server
func (whsvr *WebhookServer) serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		glog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		glog.Errorf("Can't decode body: %v", err)
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		fmt.Println(r.URL.Path)
		if r.URL.Path == "/mutate" {
			admissionResponse = whsvr.mutate(&ar)
		} else if r.URL.Path == "/validate" {
			admissionResponse = whsvr.validate(&ar)
		}
	}

	admissionReview := v1beta1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		glog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	glog.Infof("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		glog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
