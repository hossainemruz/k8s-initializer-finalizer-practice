package v1alpha1

import (
	scheme "crd-controller/pkg/client/clientset/versioned/scheme"

	v1alpha1 "crd-controller/pkg/apis/crd.emruz.com/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// CustomDeploymentsGetter has a method to return a CustomDeploymentInterface.
// A group's client should implement this interface.
type CustomDeploymentsGetter interface {
	CustomDeployments(namespace string) CustomDeploymentInterface
}

// CustomDeploymentInterface has methods to work with CustomDeployment resources.
type CustomDeploymentInterface interface {
	Create(*v1alpha1.CustomDeployment) (*v1alpha1.CustomDeployment, error)
	Update(*v1alpha1.CustomDeployment) (*v1alpha1.CustomDeployment, error)
	UpdateStatus(*v1alpha1.CustomDeployment) (*v1alpha1.CustomDeployment, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.CustomDeployment, error)
	List(opts v1.ListOptions) (*v1alpha1.CustomDeploymentList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CustomDeployment, err error)
	CustomDeploymentExpansion
}

// customDeployments implements CustomDeploymentInterface
type customDeployments struct {
	client rest.Interface
	ns     string
}

// newCustomDeployments returns a CustomDeployments
func newCustomDeployments(c *CrdV1alpha1Client, namespace string) *customDeployments {
	return &customDeployments{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the customDeployment, and returns the corresponding customDeployment object, and an error if there is any.
func (c *customDeployments) Get(name string, options v1.GetOptions) (result *v1alpha1.CustomDeployment, err error) {
	result = &v1alpha1.CustomDeployment{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("customdeployments").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CustomDeployments that match those selectors.
func (c *customDeployments) List(opts v1.ListOptions) (result *v1alpha1.CustomDeploymentList, err error) {
	result = &v1alpha1.CustomDeploymentList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("customdeployments").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested customDeployments.
func (c *customDeployments) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("customdeployments").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a customDeployment and creates it.  Returns the server's representation of the customDeployment, and an error, if there is any.
func (c *customDeployments) Create(customDeployment *v1alpha1.CustomDeployment) (result *v1alpha1.CustomDeployment, err error) {
	result = &v1alpha1.CustomDeployment{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("customdeployments").
		Body(customDeployment).
		Do().
		Into(result)
	return
}

// Update takes the representation of a customDeployment and updates it. Returns the server's representation of the customDeployment, and an error, if there is any.
func (c *customDeployments) Update(customDeployment *v1alpha1.CustomDeployment) (result *v1alpha1.CustomDeployment, err error) {
	result = &v1alpha1.CustomDeployment{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("customdeployments").
		Name(customDeployment.Name).
		Body(customDeployment).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *customDeployments) UpdateStatus(customDeployment *v1alpha1.CustomDeployment) (result *v1alpha1.CustomDeployment, err error) {
	result = &v1alpha1.CustomDeployment{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("customdeployments").
		Name(customDeployment.Name).
		SubResource("status").
		Body(customDeployment).
		Do().
		Into(result)
	return
}

// Delete takes name of the customDeployment and deletes it. Returns an error if one occurs.
func (c *customDeployments) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("customdeployments").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *customDeployments) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("customdeployments").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched customDeployment.
func (c *customDeployments) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CustomDeployment, err error) {
	result = &v1alpha1.CustomDeployment{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("customdeployments").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
