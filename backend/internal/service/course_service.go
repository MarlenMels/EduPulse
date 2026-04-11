package service

import (
	"context"

	"edupulse/internal/repo"
)

type CourseService struct {
	repo *repo.CourseRepo
}

func NewCourseService(r *repo.CourseRepo) *CourseService {
	return &CourseService{repo: r}
}

type CourseWithDetails struct {
	repo.Course
	Lessons []repo.Lesson `json:"lessons"`
	Reviews []repo.Review `json:"reviews"`
}

func (s *CourseService) List(ctx context.Context, limit int) ([]CourseWithDetails, error) {
	courses, err := s.repo.List(ctx, limit)
	if err != nil {
		return nil, err
	}

	out := make([]CourseWithDetails, 0, len(courses))
	for _, c := range courses {
		lessons, err := s.repo.LessonsByCourse(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		reviews, err := s.repo.ReviewsByCourse(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, CourseWithDetails{
			Course:  c,
			Lessons: lessons,
			Reviews: reviews,
		})
	}
	return out, nil
}
