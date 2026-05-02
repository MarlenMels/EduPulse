package service

import (
	"context"
	"errors"
	"strings"

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

func (s *CourseService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *CourseService) DeleteLesson(ctx context.Context, courseID, lessonID int64) error {
	return s.repo.DeleteLesson(ctx, courseID, lessonID)
}

func (s *CourseService) AddLessonAsset(ctx context.Context, asset repo.LessonAsset) (repo.LessonAsset, error) {
	asset.Type = strings.TrimSpace(asset.Type)
	asset.URL = strings.TrimSpace(asset.URL)
	asset.OriginalFilename = strings.TrimSpace(asset.OriginalFilename)
	if asset.LessonID <= 0 {
		return repo.LessonAsset{}, errors.New("lesson_id is required")
	}
	if asset.Type != "video" && asset.Type != "file" {
		return repo.LessonAsset{}, errors.New("type must be video or file")
	}
	if asset.URL == "" {
		return repo.LessonAsset{}, errors.New("url is required")
	}
	return s.repo.CreateLessonAsset(ctx, asset)
}

func (s *CourseService) DeleteLessonAsset(ctx context.Context, lessonID, assetID int64) error {
	if lessonID <= 0 || assetID <= 0 {
		return errors.New("invalid id")
	}
	return s.repo.DeleteLessonAsset(ctx, lessonID, assetID)
}

func (s *CourseService) List(ctx context.Context, limit int, search string) ([]CourseWithLessons, error) {
	courses, err := s.repo.List(ctx, limit, search)
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
