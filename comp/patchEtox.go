package comp

// PatchPropertiesEtox component for flower patches.
// Holds information on PPP concentrations in nectar, pollen or via contact on this patch.
type PatchPropertiesEtox struct {
	PPPconcentrationNectar float64 // PPP concentration in nectar [mug/kJ]
	PPPconcentrationPollen float64 // PPP concentration in pollen [mug/g]
	PPPcontactDose         float64 // PPP concentration for contact exposure on patch [mug/bee]
}

// ResourceEtox component for flower patches.
//
// Holds information on PPP concentrations in nectar, pollen or via contact on this patch.
type ResourceEtox struct {
	PPPconcentrationNectar float64 // PPP concentration in nectar [mug/kJ]
	PPPconcentrationPollen float64 // PPP concentration in pollen [mug/g]
	PPPcontactDose         float64 // PPP concentration for contact exposure on patch [mug/bee]
}
