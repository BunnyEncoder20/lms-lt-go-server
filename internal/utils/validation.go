package utils

import (
	"go-server/internal/models"
)

// IsValidNominationStatus Helper func to validate nomination status
func IsValidNominationStatus(status models.NominationStatus) bool {
	switch status {
	case models.NomPendingManagerAssignment,
		models.NomPendingEmployeeApproval,
		models.NomEnrolled,
		models.NomPendingManagerApproval,
		models.NomDeclined,
		models.NomRejected,
		models.NomCompleted,
		models.NomAttended:
		return true
	default:
		return false
	}
}
