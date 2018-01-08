package main

import (
	// "k8s.io/code-generator"
	 "k8s-initializer-finalizer-practice/pkg/controller"

)
func main()  {
	controller.StartDeploymentController(1)

}