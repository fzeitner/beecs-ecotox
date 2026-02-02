package globals

// LarvaeEtox contains oral doses for worker and drone larvae cohorts; divided by age like in BEEHAVE.
type LarvaeEtox struct {
	WorkerCohortDose []float64 // Mean PPP oral dose per cohort.
	DroneCohortDose  []float64 // Mean PPP oral dose per cohort.
}

// InHiveEtox contains oral doses for in-hive worker and drone cohorts; divided by age like in BEEHAVE.
type InHiveEtox struct {
	WorkerCohortDose []float64 // Mean PPP oral dose per cohort.
	DroneCohortDose  []float64 // Mean PPP oral dose per cohort.
}
