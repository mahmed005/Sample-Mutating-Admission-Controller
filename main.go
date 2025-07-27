package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func handleMutate(w http.ResponseWriter, r *http.Request) {
	admissionReview := &admissionv1.AdmissionReview{}
	err := json.NewDecoder(r.Body).Decode(admissionReview)
	if err != nil {
		sendError(w, "Could not unmarshal the Request body")
		return
	}

	var fullSpec map[string]interface{}
	err = json.Unmarshal(admissionReview.Request.Object.Raw, &fullSpec)
	if err != nil {
		sendError(w, "Could not marshal the whole object")
		return
	}

	metadata, err := json.Marshal(fullSpec["metadata"])
	if err != nil {
		sendError(w, "Could not marshal the metadata field of the object")
		return
	}
	meta := metav1.ObjectMeta{}
	err = json.Unmarshal(metadata, &meta)
	if err != nil {
		sendError(w, "Could not unmarshall the metadata fields of the object")
		return
	}
	patchType := admissionv1.PatchTypeJSONPatch
	admissionResponse := &admissionv1.AdmissionReview{TypeMeta: admissionReview.TypeMeta, Response: &admissionv1.AdmissionResponse{
		UID:       admissionReview.Request.UID,
		Allowed:   true,
		PatchType: &patchType,
	}}
	patchMap := []map[string]interface{}{}

	if meta.Labels == nil {
		patchMap = append(patchMap, map[string]interface{}{
			"op":   "add",
			"path": "/metadata/labels",
			"value": map[string]string{
				"app": fmt.Sprintf("%s-%s-%s", meta.Name, admissionReview.Request.Kind.Kind, meta.Namespace),
			},
		})
	} else if _, exists := meta.Labels["app"]; !exists {
		patchMap = append(patchMap, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/app",
			"value": fmt.Sprintf("%s-%s-%s", meta.Name, admissionReview.Request.Kind.Kind, meta.Namespace),
		})
	}

	patchBytes, err := json.Marshal(patchMap)
	if err != nil {
		sendError(w, "Could not marshal the patch for the response")
		return
	}
	(*(*admissionResponse).Response).Patch = patchBytes

	respBytes, err := json.Marshal(admissionResponse)
	if err != nil {
		sendError(w, "Could not marshal the response body")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}

func sendError(w http.ResponseWriter, err string) {
	fmt.Println(err)
	out, e := json.Marshal(map[string]string{"Err": err})
	if e != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(out)
}

func main() {
	http.HandleFunc("/mutate", handleMutate)
	log.Fatal(http.ListenAndServeTLS(":8080", "/certs/webhook.crt", "/certs/webhook-key.pem", nil))
}
