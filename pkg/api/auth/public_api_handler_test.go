package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rancher/harvester/pkg/auth/jwe"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	corev1type "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	publicAuthPath = "/v1-public/auth"
)

type input struct {
	loginInput Login
	action     string
	httpMethod string
}
type output struct {
	status int
	err    error
	body   string
}

func TestAuthAPIAction(t *testing.T) {
	var testCases = []struct {
		name     string
		given    input
		expected output
	}{
		{
			name: "login with invalid http method",
			given: input{
				loginInput: Login{
					Token: "todo",
				},
				action: http.MethodGet,
			},
			expected: output{
				status: http.StatusOK,
				err:    nil,
				body:   "todo",
			},
		},
		//{
		//	name: "login with invalid action",
		//	given: input{
		//		loginInput: Login{
		//			Token: "todo",
		//		},
		//		action: http.MethodPost,
		//	},
		//	expected: output{
		//		status: http.StatusOK,
		//		err:    nil,
		//		body:   "todo",
		//	},
		//},
		//{
		//	name: "login with valid token",
		//	given: input{
		//		loginInput: Login{
		//			Token: "todo",
		//		},
		//		action: http.MethodPost,
		//	},
		//	expected: output{
		//		status: http.StatusOK,
		//		err:    nil,
		//		body:   "todo",
		//	},
		//},
		//{
		//	name: "login with invalid token",
		//	given: input{
		//		loginInput: Login{
		//			Token: "todo",
		//		},
		//		action: http.MethodPost,
		//	},
		//	expected: output{
		//		status: http.StatusUnauthorized,
		//		err:    errors.New("todo"),
		//		body:   "todo",
		//	},
		//},
		//{
		//	name: "login with valid kubeconfig",
		//	given: input{
		//		loginInput: Login{
		//			KubeConfig: "todo",
		//		},
		//		action: http.MethodPost,
		//	},
		//	expected: output{
		//		status: http.StatusUnauthorized,
		//		err:    errors.New("todo"),
		//		body:   "todo",
		//	},
		//},
		//{
		//	name: "login with invalid kubeconfig",
		//	given: input{
		//		loginInput: Login{
		//			KubeConfig: "todo",
		//		},
		//		action: http.MethodPost,
		//	},
		//	expected: output{
		//		status: http.StatusUnauthorized,
		//		err:    errors.New("todo"),
		//		body:   "todo",
		//	},
		//},
		//{
		//	name: "logout",
		//	given: input{
		//		action: http.MethodPost,
		//	},
		//	expected: output{
		//		status: http.StatusUnauthorized,
		//		err:    errors.New("todo"),
		//		body:   "todo",
		//	},
		//},
	}

	restConfig := &rest.Config{}
	clientSet := fake.NewSimpleClientset()
	secretClient := fakeSecretClient(clientSet.CoreV1().Secrets)
	tokenManager, err := jwe.NewJWETokenManager(secretClient, "todo")
	assert.Nil(t, err, "NewJWETokenManager should return no error")

	handler := PublicAPIHandler{
		secrets: secretClient,
		restConfig: restConfig,
		tokenManager: tokenManager,
	}

	for _, tc := range testCases {
		rw := httptest.NewRecorder()
		req, err := getRequest(publicAuthPath, tc.given)

		assert.Nil(t, err, "getRequest should return no error")
		handler.ServeHTTP(rw, req)
	}

}

func getRequest(baseURL string, input input) (*http.Request, error) {
	reqURL := fmt.Sprintf("%s/auth?action=%s", baseURL, input.action)
	body, err := json.Marshal(input.loginInput)
	if err != nil {
		return nil, err
	}
	return http.NewRequest(input.httpMethod, reqURL, bytes.NewReader(body))
}

type fakeSecretClient func(string) corev1type.SecretInterface

func (c fakeSecretClient) Create(secret *corev1.Secret) (*corev1.Secret, error) {
	return c(secret.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
}

func (c fakeSecretClient) Update(secret *corev1.Secret) (*corev1.Secret, error) {
	return c(secret.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
}

func (c fakeSecretClient) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return c(namespace).Delete(context.TODO(), name, *options)
}

func (c fakeSecretClient) Get(namespace, name string, options metav1.GetOptions) (*corev1.Secret, error) {
	return c(namespace).Get(context.TODO(), name, options)
}

func (c fakeSecretClient) List(namespace string, opts metav1.ListOptions) (*corev1.SecretList, error) {
	return c(namespace).List(context.TODO(), opts)
}

func (c fakeSecretClient) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c(namespace).Watch(context.TODO(), opts)
}

func (c fakeSecretClient) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (*corev1.Secret, error) {
	return c(namespace).Patch(context.TODO(), name, pt, data, metav1.PatchOptions{}, subresources...)
}