// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"context"

	"go.opencensus.io/trace"

	"github.com/hashicorp/go-multierror"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/utils/errutils"
)

type ObservabilitySyncer interface {
	Sync(context.Context, *ObservabilitySnapshot) error
}

type ObservabilitySyncers []ObservabilitySyncer

func (s ObservabilitySyncers) Sync(ctx context.Context, snapshot *ObservabilitySnapshot) error {
	var multiErr *multierror.Error
	for _, syncer := range s {
		if err := syncer.Sync(ctx, snapshot); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr.ErrorOrNil()
}

type ObservabilityEventLoop interface {
	Run(namespaces []string, opts clients.WatchOpts) (<-chan error, error)
}

type observabilityEventLoop struct {
	emitter ObservabilityEmitter
	syncer  ObservabilitySyncer
}

func NewObservabilityEventLoop(emitter ObservabilityEmitter, syncer ObservabilitySyncer) ObservabilityEventLoop {
	return &observabilityEventLoop{
		emitter: emitter,
		syncer:  syncer,
	}
}

func (el *observabilityEventLoop) Run(namespaces []string, opts clients.WatchOpts) (<-chan error, error) {
	opts = opts.WithDefaults()
	opts.Ctx = contextutils.WithLogger(opts.Ctx, "v1.event_loop")
	logger := contextutils.LoggerFrom(opts.Ctx)
	logger.Infof("event loop started")

	errs := make(chan error)

	watch, emitterErrs, err := el.emitter.Snapshots(namespaces, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "starting snapshot watch")
	}
	go errutils.AggregateErrs(opts.Ctx, errs, emitterErrs, "v1.emitter errors")
	go func() {
		// create a new context for each loop, cancel it before each loop
		var cancel context.CancelFunc = func() {}
		// use closure to allow cancel function to be updated as context changes
		defer func() { cancel() }()
		for {
			select {
			case snapshot, ok := <-watch:
				if !ok {
					return
				}
				// cancel any open watches from previous loop
				cancel()

				ctx, span := trace.StartSpan(opts.Ctx, "observability.supergloo.solo.io.EventLoopSync")
				ctx, canc := context.WithCancel(ctx)
				cancel = canc
				err := el.syncer.Sync(ctx, snapshot)
				span.End()

				if err != nil {
					select {
					case errs <- err:
					default:
						logger.Errorf("write error channel is full! could not propagate err: %v", err)
					}
				}
			case <-opts.Ctx.Done():
				return
			}
		}
	}()
	return errs, nil
}
