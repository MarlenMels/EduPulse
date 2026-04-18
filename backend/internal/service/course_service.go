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

type CourseWithLessons struct {
	repo.Course
	Lessons []repo.Lesson `json:"lessons"`
}

func (s *CourseService) CreateWithLessons(ctx context.Context, c repo.Course, lessons []repo.Lesson) (CourseWithLessons, error) {
	course, created, err := s.repo.CreateWithLessons(ctx, c, lessons)
	if err != nil {
		return CourseWithLessons{}, err
	}
	return CourseWithLessons{Course: course, Lessons: created}, nil
}

func (s *CourseService) AddLesson(ctx context.Context, l repo.Lesson) (repo.Lesson, error) {
	return s.repo.CreateLesson(ctx, l)
}

func (s *CourseService) UpdateLesson(ctx context.Context, l repo.Lesson) (repo.Lesson, error) {
	return s.repo.UpdateLesson(ctx, l)
}

func (s *CourseService) List(ctx context.Context, limit int) ([]CourseWithLessons, error) {
	courses, err := s.repo.List(ctx, limit)
	if err != nil {
		return nil, err
	}

	out := make([]CourseWithLessons, 0, len(courses))
	for _, c := range courses {
		lessons, err := s.repo.LessonsByCourse(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, CourseWithLessons{
			Course:  c,
			Lessons: lessons,
		})
	}
	return out, nil
}
