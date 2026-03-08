package observability

import (
	"go.opentelemetry.io/otel/propagation"
)

// Propagator returns a default propagator, that will connect all components with the same mechanism.
// It should be used in all places that require a propagator. Receives an optional list of additional propagators,
// that will added to the default one. Use this only for some specific cases (usually for integration with
// third-party or old systems).
func NewTextMapPropagator(additional ...propagation.TextMapPropagator) propagation.TextMapPropagator { //nolint:ireturn
	propagators := make([]propagation.TextMapPropagator, 0, len(additional)+1)
	for i := range additional {
		if additional[i] == nil {
			continue
		}

		propagators = append(propagators, additional[i])
	}

	if len(propagators) == 0 {
		return propagation.TraceContext{}
	}

	return propagation.NewCompositeTextMapPropagator(append(propagators, propagation.TraceContext{})...)
}
