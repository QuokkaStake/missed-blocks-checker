package populators

import (
	"main/pkg/constants"
	"time"

	"github.com/rs/zerolog"
)

type Populator interface {
	Populate() error
	Name() constants.PopulatorType
	Enabled() bool
}

type Wrapper struct {
	Populator Populator
	Duration  time.Duration
	Logger    zerolog.Logger
}

func NewWrapper(
	populator Populator,
	duration time.Duration,
	logger zerolog.Logger,
) *Wrapper {
	return &Wrapper{
		Duration:  duration,
		Populator: populator,
		Logger: logger.With().
			Str("component", "populator_wrapper").
			Logger(),
	}
}

func (w *Wrapper) Start() {
	if w.Duration == 0 {
		w.Logger.Info().
			Str("name", string(w.Populator.Name())).
			Msg("Populator is disabled in config")
		return
	}

	if !w.Populator.Enabled() {
		w.Logger.Info().
			Str("name", string(w.Populator.Name())).
			Msg("Populator is disabled")
		return
	}

	w.Logger.Info().
		Str("name", string(w.Populator.Name())).
		Dur("interval", w.Duration).
		Msg("Populator is enabled")

	w.TryPopulate()

	timer := time.NewTicker(w.Duration)

	for {
		select {
		case <-timer.C:
			w.TryPopulate()
		}
	}
}

func (w *Wrapper) TryPopulate() {
	if err := w.Populator.Populate(); err != nil {
		w.Logger.Error().
			Err(err).
			Str("name", string(w.Populator.Name())).
			Msg("Got error when populating data")
	} else {
		w.Logger.Debug().
			Str("name", string(w.Populator.Name())).
			Msg("Population finished successfully")
	}
}
