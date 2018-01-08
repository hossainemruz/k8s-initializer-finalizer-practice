package main

import (
	// "k8s.io/code-generator"
	 "crd-controller/pkg/controller"

)
func main()  {
	controller.StartDeploymentController(1)

}