package main

import (
	"github.com/endocode/goju"
	"github.com/golang/glog"
)

// ReportTreeCheck enhances goju.TreeCheck
// by a Report methods to glog
type ReportTreeCheck struct {
	goju.TreeCheck
}

// Report shows the results of a TreeCheck on glog with
// the appropriate level
func (r *ReportTreeCheck) Report(level glog.Level) {
	glog.V(level).Infof("Errors       : %d\n", r.ErrorHistory.Len())
	glog.V(level).Infof("Checks   true: %d\n", r.TrueCounter)
	glog.V(level).Infof("Checks  false: %d\n", r.FalseCounter)
}

// Evaluate reports true, if no errors and no false results have been found,
// and at least one true result
func (r *ReportTreeCheck) Evaluate() bool {
	return r.ErrorHistory.Len() == 0 && r.FalseCounter == 0 && r.TrueCounter > 0
}
