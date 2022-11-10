package queue

import (
	"context"
	"dguest-scheduler/pkg/generated/clientset/versioned/fake"
	"dguest-scheduler/pkg/generated/informers/externalversions"

	"dguest-scheduler/pkg/scheduler/framework"
	"k8s.io/apimachinery/pkg/runtime"
)

// NewTestQueue creates a priority queue with an empty informer factory.
func NewTestQueue(ctx context.Context, lessFn framework.LessFunc, opts ...Option) *PriorityQueue {
	return NewTestQueueWithObjects(ctx, lessFn, nil, opts...)
}

// NewTestQueueWithObjects creates a priority queue with an informer factory
// populated with the provided objects.
func NewTestQueueWithObjects(
	ctx context.Context,
	lessFn framework.LessFunc,
	objs []runtime.Object,
	opts ...Option,
) *PriorityQueue {
	//informerFactory := informers.NewSharedInformerFactory(fake.NewSimpleClientset(objs...), 0)
	informerFactory := externalversions.NewSharedInformerFactory(fake.NewSimpleClientset(objs...), 0)
	return NewTestQueueWithInformerFactory(ctx, lessFn, informerFactory, opts...)
}

func NewTestQueueWithInformerFactory(
	ctx context.Context,
	lessFn framework.LessFunc,
	schedulerInformerFactory externalversions.SharedInformerFactory,
	opts ...Option,
) *PriorityQueue {
	pq := NewPriorityQueue(lessFn, schedulerInformerFactory, opts...)
	schedulerInformerFactory.Start(ctx.Done())
	schedulerInformerFactory.WaitForCacheSync(ctx.Done())
	return pq
}
