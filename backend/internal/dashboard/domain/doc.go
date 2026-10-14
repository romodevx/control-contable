// Package domain holds the dashboard's pure rules (M6, MODULE-dashboard
// §2): the month and the trend window, the taxpayer-level status of a
// month's estimate and the trial days left. The dashboard owns no tables
// and no tax math: its use case composes other modules' ports.
package domain
