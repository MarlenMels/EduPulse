package auth

const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleTeacher = "teacher"
	RoleStudent = "student"
	RoleParent  = "parent"
)

func IsValidRole(r string) bool {
	switch r {
	case RoleAdmin, RoleManager, RoleTeacher, RoleStudent, RoleParent:
		return true
	default:
		return false
	}
}