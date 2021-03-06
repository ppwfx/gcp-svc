package metricsutil

import (
	"context"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/armon/go-metrics"
	stackdriver "github.com/google/go-metrics-stackdriver"
	"github.com/pkg/errors"
)

func NewStackDriverMetricSink(ctx context.Context, projectID string, svcName string) (c *monitoring.MetricClient, s metrics.MetricSink, err error) {
	c, err = monitoring.NewMetricClient(ctx)
	if err != nil {
		err = errors.Wrap(err, "failed to create metric client")

		return
	}

	s = stackdriver.NewSink(c, &stackdriver.Config{
		ProjectID: projectID,
		Location:  "us-east1-c",
	})

	_, err = metrics.NewGlobal(metrics.DefaultConfig(svcName), s)
	if err != nil {
		err = errors.Wrap(err, "failed to create metrics instance")

		return
	}

	return
}

func NewInMemoryMetricSink() (s metrics.MetricSink, err error) {
	s = metrics.NewInmemSink(10*time.Second, time.Minute)

	_, err = metrics.NewGlobal(metrics.DefaultConfig("dev"), s)
	if err != nil {
		err = errors.Wrap(err, "failed to create metrics instance")

		return
	}

	return
}
