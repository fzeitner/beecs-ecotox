package main

import (
	"fmt"
	"time"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark-tools/reporter"
	"github.com/mlange-42/beecs/model"
	"github.com/mlange-42/beecs/obs"
	"github.com/mlange-42/beecs/params"
)

func main() {
	app := app.New()

	p := params.Default()
	p.Termination.MaxTicks = 365
	p.ForagingPeriod = params.ForagingPeriod{ // default of BEEHAVE_ecotox uses Rothamsted2009
		Files:       []string{"foraging-period/rothamsted2009.txt"},
		Builtin:     true,
		RandomYears: false,
	}

	pe := params.DefaultEtox()
	pe.PPPApplication.Application = true
	pe.PPPApplication.FixedNectarPollenRatio = true // BEEHAVE_ecotox default is true
	pe.PPPApplication.Nursebeefix = false           // unfixed in NetLogo
	pe.PPPApplication.HSUfix = false                // unfixed in NetLogo

	start := time.Now()

	for i := 0; i < 100; i++ {
		run(app, i, &p, &pe)
	}

	dur := time.Since(start)
	fmt.Println(dur)
}

func run(app *app.App, idx int, params params.Params, paramsEtox params.ParamsEtox) {
	app = model.DefaultEtox(params, paramsEtox, app)

	app.AddSystem(&reporter.CSV{
		Observer: &obs.DebugEtox{},
		File:     fmt.Sprintf("out/beecs-%04d.csv", idx),
		Sep:      ";",
	})

	app.Run()
}
