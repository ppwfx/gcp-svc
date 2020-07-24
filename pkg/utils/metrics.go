package utils

import (
	"context"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/armon/go-metrics"
	stackdriver "github.com/google/go-metrics-stackdriver"
	"github.com/pkg/errors"
)

func NewProductionMetrics(ctx context.Context, projectID string) (c *monitoring.MetricClient, m *metrics.Metrics, err error) {
	c, err = monitoring.NewMetricClient(ctx)
	if err != nil {
		err = errors.Wrap(err, "failed to create metric client")

		return
	}

	ss := stackdriver.NewSink(c, &stackdriver.Config{
		ProjectID: projectID,
		Location:  "us-east1-c",
	})

	conf := metrics.DefaultConfig("go-metrics-stackdriver")
	conf.EnableHostname = false

	m, err = metrics.New(conf, ss)
	if err != nil {
		err = errors.Wrap(err, "failed to create metrics instance")

		return
	}

	return
}
