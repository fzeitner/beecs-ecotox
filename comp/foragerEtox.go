package comp

// EtoxLoad component for forager squadrons.
type EtoxLoad struct {
	PPPLoad float64 // Current amount of PPP in the load [µg]
}

// PPP exposure for forager squadrons.
type PPPExpo struct {
	OralDose    float64 // Current daily oral dose of this squadron to PPP used in dose-respnse of BEEHAVE_ecotox [µg]
	ContactDose float64 // Current daily contact dose of this squadron to PPP used in dose-respnse of BEEHAVE_ecotox [µg]

	RdmSurvivalContact float64 // Survival chance or "resilience" of the squadron to PPP contact exposure
	RdmSurvivalOral    float64 // Survival chance or "resilience" of the squadron to PPP oral exposure
}
