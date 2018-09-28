/*
Copyright 2018 The Knative Authors

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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var svcSpecTemplateFile string
var port int

var svcSpecTemplate = &servingv1alpha1.ServiceSpec{}

type HookRequest struct {
	Object *eventingv1alpha1.Source `json:"object,omitempty"`
}

type HookResponse struct {
	Attachments []runtime.Object `json:"attachments"`
}

func init() {
	flag.IntVar(&port, "port", 80, "The port to listen on")
	flag.StringVar(&svcSpecTemplateFile, "service-template", "/templates/service.yaml", "A template for a Service Spec")
}

func reconcileSource(src *eventingv1alpha1.Source) ([]runtime.Object, error) {
	svc := &servingv1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "serving.knative.dev/v1alpha1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: src.Name,
		},
		Spec: *svcSpecTemplate.DeepCopy(),
	}

	channel := &eventingv1alpha1.Channel{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "eventing.knative.dev/v1alpha1",
			Kind:       "Channel",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: src.Name,
		},
		Spec: eventingv1alpha1.ChannelSpec{
			Provisioner: &eventingv1alpha1.ProvisionerReference{
				Ref: &corev1.ObjectReference{
					Name: "default",
				},
			},
		},
	}

	svc.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env = append(
		svc.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env,
		corev1.EnvVar{
			Name:  "CHANNEL_ADDRESS",
			Value: "TODO channel",
		},
	)

	attachments := []runtime.Object{
		svc,
		channel,
	}
	return attachments, nil
}

func main() {
	flag.Parse()

	file, err := os.Open(svcSpecTemplateFile)
	if err != nil {
		log.Fatal(err)
	}

	decoder := yaml.NewYAMLOrJSONDecoder(file, 10)
	if err := decoder.Decode(svcSpecTemplate); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Can't read body", http.StatusInternalServerError)
			return
		}

		var req HookRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, fmt.Sprintf("Error decoding body: %v", err), http.StatusInternalServerError)
			return
		}

		if req.Object == nil {
			http.Error(w, `Missing required key "object"`, http.StatusInternalServerError)
			return
		}

		src := req.Object
		if src.Spec.Provisioner.Ref.Name == "k8sevents" {
			attachments, err := reconcileSource(src)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error reconciling source %v: %v", src.Name, err), http.StatusInternalServerError)
				return
			}
			resp := HookResponse{
				Attachments: attachments,
			}
			if js, err := json.Marshal(resp); err == nil {
				log.Printf("resp: %s", string(js))
				fmt.Fprint(w, js)
				return
			}
		} else {
			log.Printf("Don't care about source %v because its provisioner is %v", src.Name, src.Spec.Provisioner.Ref.Name)
		}

		fmt.Fprintf(w, "{}")
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	// Shutdown on SIGTERM.
	sigchan := make(chan os.Signal, 2)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)
	sig := <-sigchan
	log.Printf("Received %v signal. Shutting down...", sig)
	server.Shutdown(context.Background())
}
