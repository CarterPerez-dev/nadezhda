// ©AngelaMos | 2026
// watch.go

package watch

import (
	"context"
	"fmt"
	"io"
	"time"
)

type Ticker interface {
	C() <-chan time.Time
	Stop()
}

type realTicker struct{ t *time.Ticker }

func (r realTicker) C() <-chan time.Time { return r.t.C }
func (r realTicker) Stop()               { r.t.Stop() }

func NewRealTicker(d time.Duration) Ticker { return realTicker{t: time.NewTicker(d)} }

type Options struct {
	Interval   time.Duration
	RunAtStart bool
	Cycle      func(context.Context) (Report, error)
	Notifier   Notifier
	NewTicker  func(time.Duration) Ticker
	Out        io.Writer
}

func (o Options) validate() error {
	if o.Cycle == nil {
		return fmt.Errorf("watch: Cycle is required")
	}
	if o.Interval <= 0 {
		return fmt.Errorf("watch: Interval must be > 0, got %v", o.Interval)
	}
	return nil
}

func Run(ctx context.Context, opts Options) error {
	if err := opts.validate(); err != nil {
		return err
	}
	out := resolveOut(opts.Out)
	newTicker := opts.NewTicker
	if newTicker == nil {
		newTicker = NewRealTicker
	}

	fmt.Fprintf(out, "watch: starting, interval %s\n", opts.Interval)

	if opts.RunAtStart {
		if stop, err := cycleAndNotify(ctx, opts, out); stop {
			return shutdown(out)
		} else if err != nil {
			fmt.Fprintf(out, "watch: cycle error: %v\n", err)
		}
	}

	ticker := newTicker(opts.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return shutdown(out)
		case <-ticker.C():
			if stop, err := cycleAndNotify(ctx, opts, out); stop {
				return shutdown(out)
			} else if err != nil {
				fmt.Fprintf(out, "watch: cycle error: %v\n", err)
			}
		}
	}
}

func Once(ctx context.Context, opts Options) error {
	if err := opts.validate(); err != nil {
		return err
	}
	out := resolveOut(opts.Out)
	_, err := cycleAndNotify(ctx, opts, out)
	return err
}

func cycleAndNotify(ctx context.Context, opts Options, out io.Writer) (stop bool, err error) {
	if ctx.Err() != nil {
		return true, nil
	}
	report, err := opts.Cycle(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return true, nil
		}
		return false, err
	}
	logReport(out, report)
	if opts.Notifier != nil && len(report.Notable) > 0 {
		if nerr := opts.Notifier.Notify(ctx, report); nerr != nil {
			fmt.Fprintf(out, "watch: notify error: %v\n", nerr)
		} else {
			fmt.Fprintf(out, "watch: notified %d notable\n", len(report.Notable))
		}
	}
	return false, nil
}

func shutdown(out io.Writer) error {
	fmt.Fprintln(out, "watch: shutdown")
	return nil
}

func resolveOut(out io.Writer) io.Writer {
	if out != nil {
		return out
	}
	return io.Discard
}

func logReport(out io.Writer, r Report) {
	fmt.Fprintf(out, "watch: cycle done in %s: %d new, %d dup, %d clusters, %d enriched (%d KEV), %d failed, %d notable\n",
		r.Duration.Round(time.Millisecond), r.NewArticles, r.Duplicates, r.Clusters, r.Enriched, r.KEVHits, r.Failed, len(r.Notable))
}
