package scpProcessor

import (
	"strconv"
)

const (
	Safe = "Safe"
	Euclid = "Euclid"
	Keter = "Keter"
	Thaumiel = "Thaumiel"
	Neutralized = "Neutralized"
	Decommissioned = "Decommissioned"
	Apollyon = "Apollyon"
	Archon = "Archon"
	Undefined = "Undefined"
)

// An instance of an SCP report
type SCPEntity struct {
	ItemNumber int
	ObjectClass	string
	SpecialContainmentProceedures string
	Description string
	URL string
}

// Converts an SCPEntity struct into a JSON string
func (scp SCPEntity) String() string {
	header := "\n====================================================================================\n"
	divider := "\n------------------------------------------------------------------------------------\n"
	itemNum := "Item Number: #" + strconv.Itoa(scp.ItemNumber)
	objCls := "Object Class: " + scp.ObjectClass
	sCp := "Special Containment Proceedures: " + scp.SpecialContainmentProceedures
	desc := "Description: " + scp.Description
	url := "URL: " + scp.URL

	return header +
		   itemNum + divider +
		   objCls + divider +
		   url + divider +
		   sCp + divider +
		   desc +
		   header
}