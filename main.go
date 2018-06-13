package main

import (
	"context"
	"encoding/base64"
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

//Environment Variables
type Specification struct {
	AWS_ACCESS_KEY string `required:false`
	AWS_SECRET_KEY string `required:false`
}

type AppSecrets struct {
	SecretMap map[string]interface{}
}

// func (s *Specification) unPackSecrets() {
// 	for {
// 		var secrets AppSecrets
// 		unpackedSecrets := s.awssecrets(secrets, "testing-stage")

// 		fmt.Println(unpackedSecrets.SecretMap)
// 		time.Sleep(60 * time.Second)
// 	}
// 	return
// }

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
		namespaceRaw := fmt.Sprintf("name=%s\n", *ns.Metadata.Name)
		namespace := strings.Trim(namespaceRaw, "name=")

		secretName := fmt.Sprintf("%sSecret", namespace)
		var secrets AppSecrets
		secrets = s.awssecrets(secrets, namespace)

		fmt.Println("Secrets map is: ")
		fmt.Println(secrets)
		// test if secrets are base64 encoded or not
		for k, v := range secrets.SecretMap {
			value := fmt.Sprintf("%s", v)
			if IsBase64(value) {
				fmt.Println("already base64 encoded")
			} else {
				fmt.Println("base64 encoding value")
				str, _ := v.(string)
				valueByte := []byte(str)
				v := EncodeBase64(valueByte)
				secrets.SecretMap[k] = []byte(v)
			}
			fmt.Println("about to call createSecret func")
		}
		createSecret(client, namespace, secretName, secrets)
	}
	time.Sleep(60 * time.Second)
}

func IsBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func EncodeBase64(s []byte) []byte {
	encoded := base64.StdEncoding.EncodeToString([]byte(s))

	return []byte(encoded)
}

func createSecret(client *k8s.Client, namespace string, name string, v AppSecrets) error {

	fmt.Println("In create secret func")
	//Convert SecretMap from map[string]interface{} to map[string][]byte
	SecretMap := v.SecretMap

	SecretMapByte := make(map[string][]byte)
	for key, value := range SecretMap {
		switch value := value.(type) {
		case string:
			SecretMap[key] = value
		}
	}

	sm := &corev1.Secret{
		Metadata: &metav1.ObjectMeta{
			Name:      &name,
			Namespace: &namespace,
		},
		Data: SecretMapByte,
	}

	err := client.Create(context.TODO(), sm)

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

	s.listNameSpaces()

	//Gather secrets
	//s.unPackSecrets()

}
