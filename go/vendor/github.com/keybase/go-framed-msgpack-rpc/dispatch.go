package rpc

import (
	"io"

	"golang.org/x/net/context"
)

type dispatcher interface {
	Call(ctx context.Context, name string, arg interface{}, res interface{}, u ErrorUnwrapper) error
	Notify(ctx context.Context, name string, arg interface{}) error
	Close()
}

type dispatch struct {
	writer encoder
	calls  *callContainer

	// Stops all loops when closed
	stopCh chan struct{}
	// Closed once all loops are finished
	closedCh chan struct{}

	log LogInterface
}

func newDispatch(enc encoder, calls *callContainer, l LogInterface) *dispatch {
	d := &dispatch{
		writer:   enc,
		calls:    calls,
		stopCh:   make(chan struct{}),
		closedCh: make(chan struct{}),

		log: l,
	}
	return d
}

func (d *dispatch) Call(ctx context.Context, name string, arg interface{}, res interface{}, u ErrorUnwrapper) error {
	profiler := d.log.StartProfiler("call %s", name)
	defer profiler.Stop()

	c := d.calls.NewCall(ctx, name, arg, res, u)

	// Have to add call before encoding otherwise we'll race the response
	d.calls.AddCall(c)
	defer d.calls.RemoveCall(c.seqid)
	errCh := d.writer.Encode([]interface{}{MethodCall, c.seqid, c.method, c.arg})

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-c.ctx.Done():
		return d.handleCancel(c)
	case <-d.stopCh:
		return io.EOF
	}

	d.log.ClientCall(c.seqid, c.method, c.arg)

	select {
	case res := <-c.resultCh:
		d.log.ClientReply(c.seqid, c.method, res.Err(), res.Res())
		return res.Err()
	case <-c.ctx.Done():
		return d.handleCancel(c)
	case <-d.stopCh:
		return io.EOF
	}
}

func (d *dispatch) Notify(ctx context.Context, name string, arg interface{}) error {
	errCh := d.writer.Encode([]interface{}{MethodNotify, name, arg})
	select {
	case err := <-errCh:
		if err == nil {
			d.log.ClientNotify(name, arg)
		}
		return err
	case <-d.stopCh:
		return io.EOF
	case <-ctx.Done():
		d.log.ClientCancel(-1, name, nil)
		return newCanceledError(name, -1)
	}
}

func (d *dispatch) Close() {
	close(d.stopCh)
}

func (d *dispatch) handleCancel(c *call) error {
	d.log.ClientCancel(c.seqid, c.method, nil)
	errCh := d.writer.Encode([]interface{}{MethodCancel, c.seqid, c.method})
	select {
	case err := <-errCh:
		if err != nil {
			d.log.Info("error while dispatching cancellation: %+v", err.Error())
		}
	default:
		// Don't block on receiving the error from the Encode
	}
	return newCanceledError(c.method, c.seqid)
}
