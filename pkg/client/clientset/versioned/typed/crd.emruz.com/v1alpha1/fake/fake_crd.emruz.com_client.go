package fake

import (
	v1alpha1 "crd-controller/pkg/client/clientset/versioned/typed/crd.emruz.com/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeCrdV1alpha1 struct {
	*testing.Fake
}

func (c *FakeCrdV1alpha1) CustomDeployments(namespace string) v1alpha1.CustomDeploymentInterface {
	return &FakeCustomDeployments{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeCrdV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
