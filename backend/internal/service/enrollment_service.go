package service

import (
	"context"
	"errors"

	"edupulse/internal/auth"
	"edupulse/internal/repo"
)

// Mutating operations are gated by RBAC at the handler layer.
type EnrollmentService struct {
	courses     *repo.CourseRepo
	users       *repo.UserRepo
	teachers    *repo.CourseTeachersRepo
	enrollments *repo.EnrollmentsRepo
	audit       *AuditService
}

func NewEnrollmentService(
	courses *repo.CourseRepo,
	users *repo.UserRepo,
	teachers *repo.CourseTeachersRepo,
	enrollments *repo.EnrollmentsRepo,
	audit *AuditService,
) *EnrollmentService {
	return &EnrollmentService{
		courses:     courses,
		users:       users,
		teachers:    teachers,
		enrollments: enrollments,
		audit:       audit,
	}
}

func (s *EnrollmentService) AddTeacher(ctx context.Context, actorID, courseID, teacherID int64) error {
	if err := s.requireCourse(ctx, courseID); err != nil {
		return err
	}
	u, err := s.users.GetByID(ctx, teacherID)
	if err != nil {
		return err
	}
	if u == nil {
		return errors.New("teacher not found")
	}
	if u.Role != auth.RoleTeacher {
		return errors.New("user is not a teacher")
	}
	if err := s.teachers.Add(ctx, courseID, teacherID); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, actorID, "course_add_teacher", "course", courseID, map[string]any{
		"teacher_id": teacherID,
	})
	return nil
}

func (s *EnrollmentService) RemoveTeacher(ctx context.Context, actorID, courseID, teacherID int64) error {
	if err := s.requireCourse(ctx, courseID); err != nil {
		return err
	}
	if err := s.teachers.Remove(ctx, courseID, teacherID); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, actorID, "course_remove_teacher", "course", courseID, map[string]any{
		"teacher_id": teacherID,
	})
	return nil
}

func (s *EnrollmentService) TeachersByCourse(ctx context.Context, courseID int64) ([]repo.User, error) {
	return s.teachers.TeachersByCourse(ctx, courseID)
}

func (s *EnrollmentService) EnrollStudent(ctx context.Context, actorID, courseID, studentID int64) error {
	if err := s.requireCourse(ctx, courseID); err != nil {
		return err
	}
	u, err := s.users.GetByID(ctx, studentID)
	if err != nil {
		return err
	}
	if u == nil {
		return errors.New("student not found")
	}
	if u.Role != auth.RoleStudent {
		return errors.New("user is not a student")
	}
	if err := s.enrollments.Enroll(ctx, courseID, studentID, actorID); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, actorID, "course_enroll", "course", courseID, map[string]any{
		"student_id": studentID,
	})
	return nil
}

func (s *EnrollmentService) UnenrollStudent(ctx context.Context, actorID, courseID, studentID int64) error {
	if err := s.requireCourse(ctx, courseID); err != nil {
		return err
	}
	if err := s.enrollments.Unenroll(ctx, courseID, studentID); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, actorID, "course_unenroll", "course", courseID, map[string]any{
		"student_id": studentID,
	})
	return nil
}

func (s *EnrollmentService) StudentsByCourse(ctx context.Context, courseID int64) ([]repo.User, error) {
	return s.enrollments.StudentsByCourse(ctx, courseID)
}

// StudentsByTeacher returns the distinct list of students across all courses a teacher teaches.
func (s *EnrollmentService) StudentsByTeacher(ctx context.Context, teacherID int64) ([]repo.User, error) {
	return s.enrollments.StudentsByTeacher(ctx, teacherID)
}

func (s *EnrollmentService) requireCourse(ctx context.Context, courseID int64) error {
	if courseID <= 0 {
		return errors.New("invalid course id")
	}
	c, err := s.courses.GetByID(ctx, courseID)
	if err != nil {
		return err
	}
	if c == nil {
		return errors.New("course not found")
	}
	return nil
}
