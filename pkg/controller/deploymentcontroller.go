package controller

import (

	clientversioned "crd-controller/pkg/client/clientset/versioned"
	crdv1alpha1 "crd-controller/pkg/apis/crd.emruz.com/v1alpha1"
	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/client-go/util/homedir"
	rt "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/clientcmd"

	"log"
	"fmt"
	"time"
	"os"
	"math/rand"
	"strconv"

)



type Controller struct{
	// for custom deployment
	clientset				 clientversioned.Clientset
	deploymentIndexer		 cache.Indexer
	deploymentInformer 		 cache.Controller
	deploymentWorkQueue		 workqueue.RateLimitingInterface
	deletedDeploymentIndexer cache.Indexer		// if deployment is deleted we may need deplyment object for further processing.

	// for pods under this custom deployment
	kubeclient 			kubernetes.Clientset
	podIndexer			cache.Indexer
	podInformer 		cache.Controller
	podWorkQueue		workqueue.RateLimitingInterface
	deletedPodIndexer 	cache.Indexer
	podLabel string

	PreviousPodPhase map[string]string
	PodOwnerKey map[string]string
	}

func NewController(clientset clientversioned.Clientset, kubeclientset kubernetes.Clientset)  *Controller{

	//---- ----------For Deployment----------

	// create ListWatcher for custom deployment
	deploymentListWatcher:=&cache.ListWatch{
		ListFunc: func(options meta_v1.ListOptions) (rt.Object, error) {
			return clientset.CrdV1alpha1().CustomDeployments(api_v1.NamespaceDefault).List(meta_v1.ListOptions{})
		},
		WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
			return clientset.CrdV1alpha1().CustomDeployments(api_v1.NamespaceDefault).Watch(options)
		},
	}

	//create workqueue for custom deployment
	deploymentWorkQueue:=workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	//create deleted indexer for custom deployment
	deletedDeploymentIndexer:=cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc,cache.Indexers{})

	//create indexer and informer for custom deployment
	deploymentIndexer,deploymentInformer:= cache.NewIndexerInformer(deploymentListWatcher, &crdv1alpha1.CustomDeployment{},0,cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key,err:=cache.MetaNamespaceKeyFunc(obj)
			if err==nil{
				deploymentWorkQueue.Add(key)
				deletedDeploymentIndexer.Delete(obj) //object is in workqueue hence it should not be here
			}

		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldDeployment:=oldObj.(*crdv1alpha1.CustomDeployment)
			newDeployment:=newObj.(*crdv1alpha1.CustomDeployment)

			if oldDeployment!=newDeployment{		//deployment has been updated
					key,err:=cache.MetaNamespaceKeyFunc(newObj)
					if err==nil{
						deploymentWorkQueue.Add(key)
					}
			}

		},
		DeleteFunc: func(obj interface{}) {
			key,err:=cache.MetaNamespaceKeyFunc(obj)
			if err==nil{
				deletedDeploymentIndexer.Add(obj)	// deployment has been deleted hence we are storing its object in case we need
				deploymentWorkQueue.Add(key)
			}
		},
	},cache.Indexers{})


	//----------------------------For Pods-------------------------------
	podListWatcher:=&cache.ListWatch{
		ListFunc: func(options meta_v1.ListOptions) (rt.Object, error) {
			return kubeclientset.CoreV1().Pods(api_v1.NamespaceDefault).List(meta_v1.ListOptions{})
		},
		WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
			return kubeclientset.CoreV1().Pods(api_v1.NamespaceDefault).Watch(options)
		},
	}

	//create workqueue for custom deployment
	podWorkQueue:=workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	//create deleted indexer for custom deployment
	deletedPodIndexer:=cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc,cache.Indexers{})

	//create indexer and informer for custom deployment
	podIndexer,podInformer:= cache.NewIndexerInformer(podListWatcher, &api_v1.Pod{},0,cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key,err:=cache.MetaNamespaceKeyFunc(obj)
			if err==nil{
				podWorkQueue.Add(key)
				deletedPodIndexer.Delete(obj)
			}

		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod:=oldObj.(*api_v1.Pod)
			newPod:=newObj.(*api_v1.Pod)

			if oldPod!=newPod{		//pod has been updated
				key,err:=cache.MetaNamespaceKeyFunc(newObj)
				if err==nil{
					podWorkQueue.Add(key)
				}
			}

		},
		DeleteFunc: func(obj interface{}) {
			key,err:=cache.MetaNamespaceKeyFunc(obj)
			if err==nil{
				deletedPodIndexer.Add(obj)
				podWorkQueue.Add(key)
			}
		},
	},cache.Indexers{})

	return &Controller{
		clientset: 				 clientset,
		deploymentIndexer:		 deploymentIndexer,
		deploymentInformer:		 deploymentInformer,
		deploymentWorkQueue:	 deploymentWorkQueue,
		deletedDeploymentIndexer:deletedDeploymentIndexer,

		kubeclient:	 		kubeclientset,
		podIndexer:	 		podIndexer,
		podInformer: 		podInformer,
		podWorkQueue:		podWorkQueue,
		deletedPodIndexer:	deletedPodIndexer,
		podLabel: "",

		PreviousPodPhase: 	make(map[string]string),
		PodOwnerKey:	 	make(map[string]string),


	}

}
func getKubeConfigPath () string {

	var kubeConfigPath string

	homeDir:=homedir.HomeDir()

	if _,err:=os.Stat(homeDir+"/.kube/config");err==nil{
		kubeConfigPath=homeDir+"/.kube/config"
	}else{
		fmt.Printf("Enter kubernetes config directory: ")
		fmt.Scanf("%s",kubeConfigPath)
	}

	return kubeConfigPath
}

func StartDeploymentController(thrediness int)  {

	//get path of kubeconfig
	configPath:=getKubeConfigPath();

	//create configuration
	config,err:=clientcmd.BuildConfigFromFlags("",configPath)

	if err!=nil{
		log.Fatal("Can't crete config. Error: %v",err)
	}

	//create clientset
	kubeclientset,err:=kubernetes.NewForConfig(config)

	if err!=nil{
		log.Fatal(err)
	}

	clientset,err:=clientversioned.NewForConfig(config)
	if err!=nil{
		log.Fatal(err)
	}


	//now create controller
	controller := NewController(*clientset,*kubeclientset)

	//lets start the controller
	stopCh:=make(chan struct{})
	defer close(stopCh)

	go controller.RunController(thrediness,stopCh)
	//wait forever
	select{}
}


//runPodWathcer function will start the controller
// stopCh channel is used to send interrupt signal to stop the controller

func (c *Controller)RunController(thrediness int, stopCh chan struct{})  {
	//Don't panic if any error occurs in this function
	defer runtime.HandleCrash()

	//stop workers when we are done
	defer c.deploymentWorkQueue.ShutDown()
	defer c.podWorkQueue.ShutDown()

	log.Println("Starting informers........")
	//starting the informers
	go c.deploymentInformer.Run(stopCh)
	go c.podInformer.Run(stopCh)

	//// Wait for all involved caches to be synced, before processing items from the podWatchQueue is started
	if !cache.WaitForCacheSync(stopCh,c.deploymentInformer.HasSynced,c.podInformer.HasSynced){
		runtime.HandleError(fmt.Errorf("Wating for caches to sync..."))
		return
	}

	// continously run workers at 1 second  interval to process tasks in working queue until stopCh signal
	for i:=0;i<thrediness;i++{
		go wait.Until(c.runWorkerForDeployment,time.Second,stopCh)
		go wait.Until(c.runWorkerForPod,time.Second,stopCh)
	}

	<-stopCh		//stack here until a message appears on stopCh channel.
}

func (c *Controller)runWorkerForDeployment()  {
	for c.processNextItemFromDeploymentWorkQueue(){	//loop until all task in the queue is processed

	}
}

func (c *Controller)runWorkerForPod()  {
	for c.processNextItemFromPodWorkQueue(){	//loop until all task in the queue is processed

	}
}

//----------------------- For Deployment ------------------------------

//Process the first item from the working queue
func (c* Controller) processNextItemFromDeploymentWorkQueue() bool{
	// Get the key of the item in front of queue
	key, isEmpty:=c.deploymentWorkQueue.Get()

	if isEmpty{		//queue is empty hence time to break the loop of the caller function
		return  false
	}

	//Tell the deploymentWorkQueue that we are done with this key.This unblocks the key for other workers
	//This allow safe parallel processing because two item with same key will never be processed in parallel
	defer c.deploymentWorkQueue.Done(key)


	err:=c.performActionOnThisDeploymentKey(key.(string))

	// if any error occours we need to handle it.
	c.handleErrorForDeployment(err,key,5)

	return true
}

func (c *Controller)performActionOnThisDeploymentKey(key string) error {

	//get the object of this key from indexer
	obj,exist,err :=c.deploymentIndexer.GetByKey(key)

	if err!=nil{
		fmt.Printf("Fetching object of key: %s from indexer failed with error: %v\n",key,err)
		return err
	}

	if !exist{	//object does not exist in indexer. maybe it is deleted.
		fmt.Printf("Deployment %s is no more exist.\n",key)

		_,exist,err:=c.deletedDeploymentIndexer.GetByKey(key) //check if it is in the deleted Indexer to be confirmed it is deleted

		if err==nil && exist{
			fmt.Printf("Deployment %s has been deleted.\n",key)
			c.deletedDeploymentIndexer.Delete(key) //done with the object
		}
	}else{
		customdeployment:=obj.(*crdv1alpha1.CustomDeployment).DeepCopy()

		fmt.Println("Sync/Add/Update happed for deployment ",customdeployment.GetName())

		fmt.Printf("Required: %v | Available: %v | Creating: %v	|	Terminating: %v\n",customdeployment.Spec.Replicas,customdeployment.Status.AvailableReplicas,customdeployment.Status.CreatingReplicas,customdeployment.Status.TerminatingReplicas)

		label:=""

		//If Current State is not same as Expected State preform necessary modification to meet the Goal.
		if customdeployment.Status.AvailableReplicas+customdeployment.Status.CreatingReplicas<customdeployment.Spec.Replicas{

			//create pod
			pod,err:= c.CreateNewPod(customdeployment.Spec.Template, customdeployment)

			//Failed to create to pod
			if err!=nil{
				fmt.Printf("Can't create pod. Reason: %v\n",err.Error())
				return err
			}

			// Pod successfully created.
			podName:=string(pod.GetName())
			c.PodOwnerKey[podName]=key
			c.PreviousPodPhase[podName]="Creating"

			mp:=pod.GetLabels()
			for key,value:= range mp{
				label+=key+"="+value
			}

			c.podLabel=label

			fmt.Printf("+++++Pod created. PodName: %s OwnerKey: %s	Label: %v\n",podName,key,label)
			err2:=c.UpdateDeploymentStatus(customdeployment)

			if err2!=nil{
				fmt.Printf("Pod created but failed to update DeploymentStatus.")
				return err2
			}

		}else if customdeployment.Status.AvailableReplicas+customdeployment.Status.CreatingReplicas> customdeployment.Spec.Replicas{

			err:=c.DeletePod(customdeployment.Status.AvailableReplicas+customdeployment.Status.CreatingReplicas-customdeployment.Spec.Replicas)

			if err!=nil{
				fmt.Println("Can't Delete Pod. Reason: ",err)
				return err
			}

			err=c.UpdateDeploymentStatus(customdeployment)
			if err!=nil{
				fmt.Println("Failed to update DeploymentStatus.")
				return err
			}

		}else{
			// Everything is ok. nothing to do. :)
		}
	}
	return  nil
}

func (c *Controller)handleErrorForDeployment(err error, key interface{},maxNumberOfRetry int)  {
	if err==nil{
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.deploymentWorkQueue.Forget(key)
		return
	}

	//Requeue the key for retry if it does not exceed the maximum limit of retry
	if c.deploymentWorkQueue.NumRequeues(key)<maxNumberOfRetry{

		fmt.Printf("Error in processing event with key: %s\nError: %v",key,err.Error())
		fmt.Printf("Retraying to process the event for %s\n",key)

		// Requeuing the key for retry. This will increase NumRequeues for this key.
		c.deploymentWorkQueue.AddRateLimited(key)
		return
	}

	//Maximum number of requeue limit is over. Forget about the key.
	fmt.Printf("Can't process event with key:%s . The key is being dropped.\n",key)
	c.deploymentWorkQueue.Forget(key)
	runtime.HandleError(err)
}




//----------------------- For Pods ------------------------------

//Process the first item from the working queue
func (c* Controller) processNextItemFromPodWorkQueue() bool{
	// Get the key of the item in front of queue
	key, isEmpty:=c.podWorkQueue.Get()

	if isEmpty{		//queue is empty hence time to break the loop of the caller function
		return  false
	}

	//Tell the podWorkQueue that we are done with this key.This unblocks the key for other workers
	//This allow safe parallel processing because two item with same key will never be processed in parallel
	defer c.podWorkQueue.Done(key)


	err:=c.performActionOnThisPodKey(key.(string))

	// if any error occours we need to handle it.
	c.handleErrorForPod(err,key,5)

	return true
}

func (c *Controller)performActionOnThisPodKey(key string) error {

	//get the object of this key from indexer
	obj,exist,err :=c.podIndexer.GetByKey(key)

	if err!=nil{
		fmt.Printf("Fetching object of key: %s from indexer failed with error: %v\n",key,err)
		return err
	}

	if !exist{	//object is not exist in indexer. maybe it is deleted.

		fmt.Printf("Pod %s is no more exist.\n",key)
		deletedObj,exist,err:=c.deletedPodIndexer.GetByKey(key) //check if it is in the deleted Indexer to be confirmed it is deleted

		if err==nil && exist{
			fmt.Printf("pod %s has been deleted.\n",key)

			c.deletedPodIndexer.Delete(key) //done with the object
			deletedPod:=deletedObj.(*api_v1.Pod).DeepCopy()

			podPhase:=deletedPod.Status.Phase
			podName:=deletedPod.GetName()

			if podPhase=="Succeeded" || podPhase=="Failed"{
				c.PreviousPodPhase[podName]="Terminated"
			}
			podownerkey:=c.PodOwnerKey[podName]
			ownerObj,exist,err :=c.deploymentIndexer.GetByKey(podownerkey)

			if err!=nil{
				fmt.Println("Can't get podOwner object. Reason: ",err)
				return err
			}

			if !exist{
				fmt.Println("Owner does not exist")
				return err
			}

			customdeployment:=ownerObj.(*crdv1alpha1.CustomDeployment).DeepCopy()

			err =c.UpdateDeploymentStatus(customdeployment)

			return err

		}


	}else{
		pod:=obj.(*api_v1.Pod).DeepCopy()

		fmt.Println("Sync/Add/Update happed for Pod: ",pod.GetName())
		curPodPhase:=pod.Status.Phase
		podName:=pod.GetName()

		if c.PreviousPodPhase[podName]=="Creating"&&curPodPhase=="Running"{
			c.PreviousPodPhase[podName]="Running"
		}else{
			// no change required
		}

		podownerkey:=c.PodOwnerKey[podName]
		ownerObj,exist,err :=c.deploymentIndexer.GetByKey(podownerkey)

		if err!=nil{
			fmt.Println("Can't get podOwner object. Reason: ",err)
			return err
		}

		if !exist{
			fmt.Println("Owner does not exist")
			return err
		}

		customdeployment:=ownerObj.(*crdv1alpha1.CustomDeployment).DeepCopy()

		err =c.UpdateDeploymentStatus(customdeployment)
		if err!=nil{
			fmt.Println("Can't update DeploymentStatus. Reason: ",err)
			return err
		}
		fmt.Printf("Required: %v | Available: %v | Creating: %v	|	Terminating: %v\n",customdeployment.Spec.Replicas,customdeployment.Status.AvailableReplicas,customdeployment.Status.CreatingReplicas,customdeployment.Status.TerminatingReplicas)

		return err

	}
	return  nil
}

func (c *Controller)handleErrorForPod(err error, key interface{},maxNumberOfRetry int)  {
	if err==nil{
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.podWorkQueue.Forget(key)
		return
	}

	//Requeue the key for retry if it does not exceed the maximum limit of retry
	if c.podWorkQueue.NumRequeues(key)<maxNumberOfRetry{

		fmt.Printf("Error in processing event with key: %s\nError: %v",key,err.Error())
		fmt.Printf("Retraying to process the event for %s\n",key)

		// Requeuing the key for retry. This will increase NumRequeues for this key.
		c.podWorkQueue.AddRateLimited(key)
		return
	}

	//Maximum number of requeue limit is exceeded. Forget about the key.
	//Maximum number of requeue limit is exceeded. Forget about the key.
	fmt.Printf("Can't process event with key:%s . The key is being dropped.\n",key)
	c.podWorkQueue.Forget(key)
	runtime.HandleError(err)
}


func (c *Controller)CreateNewPod(podTemplate crdv1alpha1.CustomPodTemplate, customdeployment *crdv1alpha1.CustomDeployment)  (*api_v1.Pod,error){

	podClient:=c.kubeclient.CoreV1().Pods(api_v1.NamespaceDefault)

	pod:=&api_v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:	customdeployment.GetName()+"-"+strconv.Itoa(rand.Int()),
			Labels: podTemplate.GetObjectMeta().GetLabels(),
		},
		Spec: podTemplate.Spec,
	}

	newPod,err:= podClient.Create(pod)

	if err==nil{
		fmt.Printf("New pod with name %v has been created.\n",newPod.GetName())
	}

	return newPod,err
}

func (c *Controller)DeletePod(deletionLimit int32)  error{

	podClient:= c.kubeclient.CoreV1().Pods(api_v1.NamespaceDefault)

	podList,err:=c.kubeclient.CoreV1().Pods(api_v1.NamespaceDefault).List(meta_v1.ListOptions{LabelSelector: c.podLabel})
	if err!=nil{
		fmt.Println("Can't get pod list. Reason: ",err)
	}

	deletedPod:=int32(0)

	for _,pod:=range podList.Items{

		delErr:=podClient.Delete(pod.GetName(),&meta_v1.DeleteOptions{})

		if delErr!=nil{
			return delErr
		}else{
			c.PreviousPodPhase[pod.GetName()]="Terminating"
			deletedPod++
			if deletedPod>=deletionLimit{
				break
			}
		}
	}

	return nil
}

func (c *Controller)UpdateDeploymentStatus(customdeployment *crdv1alpha1.CustomDeployment) error{

	running:=0
	creating:=0
	terminating:=0

	podList,err:=c.kubeclient.CoreV1().Pods(api_v1.NamespaceDefault).List(meta_v1.ListOptions{LabelSelector: c.podLabel})
	if err!=nil{
		fmt.Println("Can't get pod list. Reason: ",err)
	}

	for _,pod:=range podList.Items{

		if c.PreviousPodPhase[pod.GetName()]=="Creating"{
			creating++
		}else if c.PreviousPodPhase[pod.GetName()]=="Running"{
			running++
		}else if c.PreviousPodPhase[pod.GetName()]=="Terminating"{
			terminating++
		}

	}

	//Don't modify cache. Work on it's copy
	customdeploymentCopy:= customdeployment.DeepCopy()

	customdeploymentCopy.Spec.Replicas = customdeployment.Spec.Replicas
	customdeploymentCopy.Status.AvailableReplicas = int32(running)
	customdeploymentCopy.Status.CreatingReplicas = int32(creating)
	customdeploymentCopy.Status.TerminatingReplicas = int32(terminating)

	//Now update the cache
	_,err=c.clientset.CrdV1alpha1().CustomDeployments(api_v1.NamespaceDefault).Update(customdeploymentCopy)

	return err
}
