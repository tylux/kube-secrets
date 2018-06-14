package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/common/log"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/kelseyhightower/envconfig"
)

//Specication Environment Variables
type Specification struct {
	AWS_ACCESS_KEY     string `required:false`
	AWS_SECRET_KEY     string `required:false`
	SecretPrefix       string `default:"kubernetes"`
	ExcludedNamespaces string `default:"default,kube-public,kube-system,docker"`
}

type AppSecrets struct {
	SecretMap map[string]interface{}
}

func (s *Specification) listNameSpaces() {
	client, err := k8s.NewInClusterClient()
	if err != nil {
		log.Fatal(err)
	}

	var ns corev1.NamespaceList
	if err := client.List(context.Background(), "", &ns); err != nil {
		log.Fatal(err)
	}

	for _, ns := range ns.Items {
		namespaceRaw := fmt.Sprintf("name=%s", *ns.Metadata.Name)
		//trim off name= from namespaceRaw
		namespace := namespaceRaw[5:]

		exclusions := strings.Split(s.ExcludedNamespaces, ",")

		if !stringInSlice(namespace, exclusions) {
			fmt.Println("Namespace: ", namespace)

			secretName := fmt.Sprintf("%s-secret", namespace)
			var secrets AppSecrets
			secrets, err = s.awssecrets(secrets, namespace)
			if err != nil {
				fmt.Println("Secret does not exist in AWS Secrets manager\n")
			} else {
				//create secret if no errors returned from AWS Secrets Manager
				createSecret(client, namespace, secretName, secrets)
			}
		} else {
			fmt.Printf("%s is part of Excluded Namespaces\n", namespace)
		}
	}
	time.Sleep(90 * time.Second)
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func createSecret(client *k8s.Client, namespace string, name string, v AppSecrets) error {

	fmt.Println("Trying to create/update secret for: ", namespace)

	//Convert SecretMap from map[string]interface{} to map[string][]byte
	SecretMap := v.SecretMap
	SecretMapByte := make(map[string][]byte)

	for k, v := range SecretMap {
		str, _ := v.(string)
		valueByte := []byte(str)
		SecretMapByte[k] = valueByte
	}

	sm := &corev1.Secret{
		Metadata: &metav1.ObjectMeta{
			Name:      &name,
			Namespace: &namespace,
		},
		Data: SecretMapByte,
	}

	err := client.Update(context.TODO(), sm)
	if err != nil {
		fmt.Println(err)
	}

	// If an HTTP error was returned by the API server, it will be of type
	// *k8s.APIError. This can be used to inspect the status code.
	if apiErr, ok := err.(*k8s.APIError); ok {
		// Resource already exists. Carry on.
		if apiErr.Code == http.StatusConflict {
			return nil
		}
	}
	return fmt.Errorf("create configmap: %v", err)
}

func main() {
	fmt.Println("Staring app...")

	//Gather environment variables that start with CI_
	var s Specification
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal(err.Error())
	}

	for {
		s.listNameSpaces()
	}

}
