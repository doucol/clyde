package dynFil

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/doucol/clyde/internal/cmdContext"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type dynFil struct{}

func (d *dynFil) Export(ctx context.Context) error {
	client := cmdContext.ClientDynFromContext(ctx)
	resource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, time.Minute, corev1.NamespaceAll, nil)
	informer := factory.ForResource(resource).Informer()

	mux := &sync.RWMutex{}
	// handle := false
	var handle cache.ResourceEventHandlerRegistration
	handle, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			log.Printf("AddFunc: av: %v, ns: %v, name: %v\n", pod.APIVersion, pod.Namespace, pod.Name)
			mux.RLock()
			defer mux.RUnlock()
			if !handle.HasSynced() {
				return
			}

			// Handler logic
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod := newObj.(*corev1.Pod)
			log.Printf("UpdateFunc: av: %v, ns: %v, name: %v\n", pod.APIVersion, pod.Namespace, pod.Name)
			mux.RLock()
			defer mux.RUnlock()
			if !handle.HasSynced() {
				return
			}

			// Handler logic
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			log.Printf("DeleteFunc: av: %v, ns: %v, name: %v\n", pod.APIVersion, pod.Namespace, pod.Name)
			mux.RLock()
			defer mux.RUnlock()
			if !handle.HasSynced() {
				return
			}

			// Handler logic
		},
	})
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go informer.Run(ctx.Done())

	isSynced := cache.WaitForCacheSync(ctx.Done(), informer.HasSynced)
	if handle.HasSynced() != isSynced {
		log.Fatal("synced state mismatch")
	}
	// mux.Lock()
	// handle = isSynced
	// mux.Unlock()

	if !isSynced {
		log.Fatal("failed to sync")
	}

	<-ctx.Done()
	return nil
}
